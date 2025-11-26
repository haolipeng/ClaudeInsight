package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target amd64 -cc clang claude_monitor claude_monitor.bpf.c -- -I. -I../../vmlinux/amd64

const (
	maxDataSize = 4096
)

// sslEvent matches the C struct ssl_event in claude_monitor.bpf.c
// Note: C struct uses __attribute__((packed)), so no padding here
type sslEvent struct {
	Pid     uint32
	DataLen uint32
	IsRead  uint8
	Comm    [16]byte
	Data    [maxDataSize]byte
}

// Node.js binary paths to try
var nodeBinaryPaths = []string{
	"/opt/node-v22.20.0/bin/node",
	"/usr/bin/node",
	"/usr/local/bin/node",
	"/opt/nodejs/bin/node",
}

// findNodeBinary finds the Node.js binary path
func findNodeBinary() (string, error) {
	for _, path := range nodeBinaryPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("Node.js binary not found in any of the expected paths")
}

// isHTTPRequest checks if data is an HTTP request
func isHTTPRequest(data []byte) bool {
	if len(data) < 16 {
		return false
	}
	methods := []string{"GET ", "POST ", "PUT ", "DELETE ", "HEAD ", "OPTIONS ", "PATCH "}
	for _, method := range methods {
		if strings.HasPrefix(string(data), method) {
			return true
		}
	}
	return false
}

// isHTTPResponse checks if data is an HTTP response
func isHTTPResponse(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	return strings.HasPrefix(string(data), "HTTP/")
}

// printData prints the captured data
func printData(event *sslEvent) {
	timestamp := time.Now().Format("15:04:05")
	direction := "⬆️  WRITE"
	if event.IsRead == 1 {
		direction = "⬇️  READ"
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Printf("%s [%s] PID: %d (%s) - %d bytes\n",
		direction, timestamp, event.Pid,
		string(bytes.TrimRight(event.Comm[:], "\x00")),
		event.DataLen)
	fmt.Println(strings.Repeat("-", 70))

	// Get actual data
	data := event.Data[:event.DataLen]

	// Check if it's HTTP
	if isHTTPRequest(data) || isHTTPResponse(data) {
		fmt.Println("📄 HTTP Data:")
		// Print first 1024 bytes
		printLen := event.DataLen
		if printLen > 1024 {
			printLen = 1024
		}
		fmt.Println(string(data[:printLen]))
		if event.DataLen > 1024 {
			fmt.Printf("... (truncated, %d more bytes)\n", event.DataLen-1024)
		}
	} else {
		// Print hex dump for non-HTTP data
		fmt.Println("🔢 Raw Data (hex):")
		printLen := event.DataLen
		if printLen > 256 {
			printLen = 256
		}
		for i := uint32(0); i < printLen; i += 16 {
			fmt.Printf("%04x  ", i)
			for j := uint32(0); j < 16 && i+j < printLen; j++ {
				fmt.Printf("%02x ", data[i+j])
			}
			fmt.Println()
		}
		if event.DataLen > 256 {
			fmt.Printf("... (truncated, showing first 256 of %d bytes)\n", event.DataLen)
		}
	}
}

func main() {
	var targetPID int
	flag.IntVar(&targetPID, "pid", 0, "Target PID to monitor (0 = all processes)")
	flag.Parse()

	// Remove memory lock limit
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal("Removing memlock:", err)
	}

	// Find Node.js binary
	nodePath, err := findNodeBinary()
	if err != nil {
		log.Fatal("Failed to find Node.js binary:", err)
	}
	fmt.Printf("📌 Found Node.js binary: %s\n", nodePath)

	// Load BPF objects
	spec, err := loadClaude_monitor()
	if err != nil {
		log.Fatal("Loading BPF spec:", err)
	}

	objs := claude_monitorObjects{}
	if err := spec.LoadAndAssign(&objs, nil); err != nil {
		var ve *ebpf.VerifierError
		if errors.As(err, &ve) {
			log.Fatalf("Verifier error: %+v", ve)
		}
		log.Fatal("Loading BPF objects:", err)
	}
	defer objs.Close()

	// Set target PID in map
	key := uint32(0)
	pidValue := uint32(targetPID)
	if err := objs.TargetPidMap.Update(&key, &pidValue, ebpf.UpdateAny); err != nil {
		log.Fatal("Updating target PID map:", err)
	}

	if targetPID > 0 {
		fmt.Printf("🎯 Monitoring PID: %d\n", targetPID)
	} else {
		fmt.Println("🌍 Monitoring all Node.js processes")
	}

	// Open Node.js executable
	ex, err := link.OpenExecutable(nodePath)
	if err != nil {
		log.Fatal("Opening Node.js executable:", err)
	}

	// Attach uprobe to SSL_write
	upSSLWrite, err := ex.Uprobe("SSL_write", objs.UprobeSslWrite, nil)
	if err != nil {
		log.Fatal("Attaching uprobe to SSL_write:", err)
	}
	defer upSSLWrite.Close()
	fmt.Println("✅ Attached uprobe to SSL_write")

	// Attach uprobe to SSL_read (entry)
	upSSLReadEntry, err := ex.Uprobe("SSL_read", objs.UprobeSslReadEntry, nil)
	if err != nil {
		log.Fatal("Attaching uprobe to SSL_read entry:", err)
	}
	defer upSSLReadEntry.Close()
	fmt.Println("✅ Attached uprobe to SSL_read (entry)")

	// Attach uretprobe to SSL_read (exit)
	upSSLReadExit, err := ex.Uretprobe("SSL_read", objs.UprobeSslReadExit, nil)
	if err != nil {
		log.Fatal("Attaching uretprobe to SSL_read exit:", err)
	}
	defer upSSLReadExit.Close()
	fmt.Println("✅ Attached uretprobe to SSL_read (exit)")

	// Open ringbuffer
	rd, err := ringbuf.NewReader(objs.Events)
	if err != nil {
		log.Fatal("Opening ringbuf reader:", err)
	}
	defer rd.Close()

	fmt.Println("\n🎯 Monitoring SSL_write() and SSL_read() calls... Press Ctrl+C to stop")
	fmt.Println("💡 Now use Claude Code to see captured HTTPS traffic\n")

	// Handle Ctrl+C
	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stopper
		fmt.Println("\n\n👋 Shutting down...")
		rd.Close()
	}()

	// Read events
	var event sslEvent
	for {
		record, err := rd.Read()
		if err != nil {
			if errors.Is(err, ringbuf.ErrClosed) {
				break
			}
			log.Printf("Reading from ringbuf: %s", err)
			continue
		}

		// Parse event
		if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &event); err != nil {
			log.Printf("Parsing event: %s", err)
			continue
		}

		// Print the captured data
		printData(&event)
	}
}
