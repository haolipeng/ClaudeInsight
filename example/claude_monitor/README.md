# Claude Code HTTPS Monitor

这是一个简化的示例程序，用于验证能否使用 eBPF uprobe 捕获 Claude Code 的 HTTPS 流量。

## 原理

Claude Code 是基于 Node.js 的应用，Node.js 静态链接了 OpenSSL。这个程序通过 eBPF uprobe 直接 hook Node.js 二进制文件中的 `SSL_read` 和 `SSL_write` 函数，从而捕获明文的 HTTPS 流量。

## 依赖要求

1. **内核版本**: Linux 4.14+ (推荐 5.x+)
2. **工具链**:
   - Go 1.21+
   - clang/llvm
   - bpftool (用于生成 vmlinux.h)
   - Linux headers

安装依赖（Ubuntu/Debian）:
```bash
sudo apt-get update
sudo apt-get install -y clang llvm libbpf-dev linux-headers-$(uname -r) bpftool
```

## 编译

```bash
# 1. 进入目录
cd /home/work/ClaudeInsight/example/claude_monitor

# 2. 下载 Go 依赖
go mod download

# 3. 编译（会自动生成 vmlinux.h 和 BPF 字节码）
make
```

## 使用方法

### 1. 查找 Claude Code 的 PID

```bash
ps aux | grep node | grep -v grep
# 或者
ps aux | grep claude | grep -v grep
```

### 2. 运行监控程序

监控所有 Node.js 进程:
```bash
sudo ./claude_monitor
```

监控特定 PID:
```bash
sudo ./claude_monitor -pid <CLAUDE_CODE_PID>
```

### 3. 触发 HTTPS 流量

在另一个终端使用 Claude Code 发送请求，你应该能看到解密后的 HTTPS 流量。

## 输出示例

```
📌 Found Node.js binary: /opt/node-v22.20.0/bin/node
✅ Attached uprobe to SSL_write
✅ Attached uprobe to SSL_read (entry)
✅ Attached uretprobe to SSL_read (exit)

🎯 Monitoring SSL_write() and SSL_read() calls... Press Ctrl+C to stop
💡 Now use Claude Code to see captured HTTPS traffic

======================================================================
⬆️  WRITE [14:23:45] PID: 98870 (node) - 1024 bytes
----------------------------------------------------------------------
📄 HTTP Data:
POST /v1/messages HTTP/1.1
Host: api.anthropic.com
Content-Type: application/json
...
```

## 故障排查

### 1. "Node.js binary not found"
确保 Node.js 已安装，或修改 `main.go` 中的 `nodeBinaryPaths` 添加你的 Node.js 路径。

### 2. "Verifier error"
- 确保内核版本 >= 4.14
- 尝试升级 Linux headers
- 检查 `/proc/sys/kernel/unprivileged_bpf_disabled` 是否为 0

### 3. "Permission denied"
必须使用 root 权限运行:
```bash
sudo ./claude_monitor
```

## 清理

```bash
make clean
```

## 文件说明

- `claude_monitor.bpf.c` - eBPF 内核态代码（C语言）
- `main.go` - 用户态代码（Go语言）
- `Makefile` - 编译脚本
- `go.mod` - Go 模块依赖

## 技术细节

1. **uprobe**: 在 Node.js 二进制文件的 SSL_read/SSL_write 函数入口设置探针
2. **uretprobe**: 在函数返回时设置探针，捕获返回值
3. **ringbuf**: 使用环形缓冲区从内核传递数据到用户空间
4. **PID 过滤**: 可以指定只监控特定进程

## 参考

- [eBPF Tutorial](https://ebpf.io/)
- [cilium/ebpf](https://github.com/cilium/ebpf)
- [BPF CO-RE](https://nakryiko.com/posts/bpf-portability-and-co-re/)
