package uprobe

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ac "claudeinsight/agent/common"
	"claudeinsight/bpf"
	"claudeinsight/common"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

const (
	kNodeBinary = "node"
)

// nodeBinaryPaths contains possible Node.js binary file paths
var nodeBinaryPaths = []string{
	"/opt/node-v22.20.0/bin/node",
	"/usr/bin/node",
	"/usr/local/bin/node",
	"/opt/nodejs/bin/node",
	"/usr/local/nodejs/bin/node",
}

// findNodeBinary finds the Node.js binary file path
// Priority:
// 1. Check if the process executable is node
// 2. Check if the process loaded libraries contain node executable
// 3. Check predefined common paths
func findNodeBinary(pid int) (string, error) {
	// 1. First check the process executable file
	execPath, err := common.GetExecutablePathFromPid(pid)
	if err == nil && isNodeExecutable(execPath) {
		common.UprobeLog.Debugf("Found Node.js binary from process executable: %s", execPath)
		return execPath, nil
	}

	// 2. Check process loaded library paths to see if loaded from node path
	nodePath := findNodeFromMaps(pid)
	if nodePath != "" {
		common.UprobeLog.Debugf("Found Node.js binary from maps: %s", nodePath)
		return nodePath, nil
	}

	// 3. Check predefined common paths
	for _, path := range nodeBinaryPaths {
		if fileExists(path) && isNodeExecutable(path) {
			common.UprobeLog.Debugf("Found Node.js binary from predefined paths: %s", path)
			return path, nil
		}
	}

	return "", fmt.Errorf("Node.js binary not found for pid %d", pid)
}

// isNodeExecutable checks if the given path is a Node.js executable file
func isNodeExecutable(path string) bool {
	// Check if filename contains "node"
	baseName := filepath.Base(path)
	if !strings.Contains(strings.ToLower(baseName), "node") {
		return false
	}

	// Check if file exists and is executable
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check if it's a regular file with execute permission
	if !info.Mode().IsRegular() {
		return false
	}

	// Check for execute permission
	if info.Mode().Perm()&0111 == 0 {
		return false
	}

	return true
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// findNodeFromMaps finds node binary file path from process memory maps
func findNodeFromMaps(pid int) string {
	paths := common.GetMapPaths(pid)
	for _, path := range paths {
		// Skip non-executable file paths
		if !strings.HasPrefix(path, "/") {
			continue
		}

		// Check if path contains node-related directory or file
		if strings.Contains(path, "/node") && isNodeExecutable(path) {
			return path
		}

		// Check if it's a symbolic link pointing to node
		if realPath, err := filepath.EvalSymlinks(path); err == nil {
			if strings.Contains(realPath, "/node") && isNodeExecutable(realPath) {
				return realPath
			}
		}
	}
	return ""
}

// IsNodeJsProcess checks if a process is a Node.js process (exported for use in loader)
func IsNodeJsProcess(pid int) bool {
	execPath, err := common.GetExecutablePathFromPid(pid)
	if err != nil {
		return false
	}

	// Check executable filename
	baseName := filepath.Base(execPath)
	if strings.Contains(strings.ToLower(baseName), "node") {
		return true
	}

	// Check process memory maps
	paths := common.GetMapPaths(pid)
	for _, path := range paths {
		// Look for Node.js-related library files
		if strings.Contains(path, "libnode.so") {
			return true
		}
		if strings.Contains(path, "/node") && strings.Contains(path, "/bin/") {
			return true
		}
	}

	return false
}

// getNodeBinaryPath gets the binary file path of a Node.js process
// Used for attaching uprobes
func getNodeBinaryPath(pid int) (string, error) {
	return findNodeBinary(pid)
}

// AttachNodeJsUprobe attaches uprobes to Node.js binary file
// Node.js typically statically links OpenSSL, so we need to attach directly to the Node.js binary
func AttachNodeJsUprobe(pid int) ([]link.Link, error) {
	nodePath, err := getNodeBinaryPath(pid)
	if err != nil {
		return nil, fmt.Errorf("failed to find Node.js binary: %w", err)
	}

	common.UprobeLog.Infof("Attaching to Node.js binary: %s", nodePath)

	// Check if already attached to this path
	attachedLibPathsMutex.Lock()
	if _, found := attachedLibPaths[nodePath]; found {
		attachedLibPathsMutex.Unlock()
		return []link.Link{}, nil
	}
	attachedLibPaths[nodePath] = true
	attachedLibPathsMutex.Unlock()

	// Try to detect OpenSSL version
	// For Node.js, we assume it uses newer OpenSSL 3.x
	// Can be determined in various ways: check Node.js version, check symbols, etc.
	versionKey := detectNodeJsOpenSSLVersion(nodePath)
	common.UprobeLog.Debugf("Node.js OpenSSL version key: %s", versionKey)

	bpfFunc, ok := sslVersionBpfMap[versionKey]
	if !ok {
		common.UprobeLog.Warnf("No BPF program for version %s, trying default", versionKey)
		// Use default 3.0 version
		bpfFunc = sslVersionBpfMap[Linuxdefaulefilename30]
	}

	sslEx, err := link.OpenExecutable(nodePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Node.js executable: %w", err)
	}

	spec, objs, err := bpfFunc()
	if err != nil {
		return nil, fmt.Errorf("failed to load BPF spec: %w", err)
	}

	collectionOptions := &ebpf.CollectionOptions{
		Programs: ebpf.ProgramOptions{
			LogSize:     10 * 1024,
			KernelTypes: ac.CollectionOpts.Programs.KernelTypes,
		},
		MapReplacements: getMapReplacementsForOpenssl(),
	}

	err = spec.LoadAndAssign(objs, collectionOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to load BPF programs: %w", err)
	}

	var l link.Link
	var links []link.Link

	// Node.js statically links OpenSSL, use nested syscall mode
	socketFDAccess := kNestedSyscall

	// SSL_read
	l, err = sslEx.Uprobe(LibSslReadFuncName, bpf.GetProgramFromObjs(objs, buildBPFFuncName(LibSslReadFuncName, false, false, socketFDAccess)), nil)
	if err != nil {
		common.UprobeLog.Warnf("Failed to attach SSL_read to Node.js: %v", err)
	} else {
		links = append(links, l)
	}

	l, err = sslEx.Uretprobe(LibSslReadFuncName, bpf.GetProgramFromObjs(objs, buildBPFFuncName(LibSslReadFuncName, false, true, socketFDAccess)), nil)
	if err != nil {
		common.UprobeLog.Warnf("Failed to attach SSL_read uretprobe to Node.js: %v", err)
	} else {
		links = append(links, l)
	}

	// SSL_write
	l, err = sslEx.Uprobe(LibSslWriteFuncName, bpf.GetProgramFromObjs(objs, buildBPFFuncName(LibSslWriteFuncName, false, false, socketFDAccess)), nil)
	if err != nil {
		common.UprobeLog.Warnf("Failed to attach SSL_write to Node.js: %v", err)
	} else {
		links = append(links, l)
	}

	l, err = sslEx.Uretprobe(LibSslWriteFuncName, bpf.GetProgramFromObjs(objs, buildBPFFuncName(LibSslWriteFuncName, false, true, socketFDAccess)), nil)
	if err != nil {
		common.UprobeLog.Warnf("Failed to attach SSL_write uretprobe to Node.js: %v", err)
	} else {
		links = append(links, l)
	}

	if len(links) == 0 {
		return nil, fmt.Errorf("failed to attach any probes to Node.js")
	}

	common.UprobeLog.Infof("Successfully attached %d probes to Node.js binary %s", len(links), nodePath)
	return links, nil
}

// detectNodeJsOpenSSLVersion detects the OpenSSL version used by Node.js
// Node.js 22.x uses OpenSSL 3.5.x
func detectNodeJsOpenSSLVersion(nodePath string) string {
	// Try to detect OpenSSL version from binary file
	versionKey, err := getOpenSslVersionKey(nodePath)
	if err == nil && versionKey != "" {
		common.UprobeLog.Debugf("Detected Node.js OpenSSL version: %s", versionKey)
		return versionKey
	}

	// Node.js 22.x defaults to OpenSSL 3.5.x
	// If unable to detect, return version 3.5
	common.UprobeLog.Debugf("Could not detect Node.js OpenSSL version, defaulting to 3.5")
	return Linuxdefaulefilename350
}
