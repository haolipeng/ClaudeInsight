# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Coding Standards

**IMPORTANT: All code comments MUST be written in English.**

When writing or modifying code in this repository:
- All comments (function comments, inline comments, TODO comments) must be in English
- Variable names, function names, and identifiers should use English words
- Log messages and error messages should be in English
- Documentation files should be in English

## Project Overview

claudeinsight is an eBPF-based network issue analysis tool for capturing and analyzing network requests (HTTP, DNS) in real-time. It provides kernel-level latency details and automatic SSL/TLS decryption.

## Build Commands

```bash
# Development build (local testing)
make build-bpf && make

# Production build with BTF support (for older kernels without BTF)
# x86_64:
make build-bpf && make btfgen BUILD_ARCH=x86_64 ARCH_BPF_NAME=x86 && make
# arm64:
make build-bpf && make btfgen BUILD_ARCH=arm64 ARCH_BPF_NAME=arm64 && make

# Run tests
make test-go

# Format code
make format-go

# Debug build (with symbols for dlv)
make ClaudeInsight-debug

# Remote debugging
make remote-debug
```

## Running ClaudeInsight

Requires root privileges:
```bash
sudo ./claudeinsight watch http           # Watch HTTP traffic
sudo ./claudeinsight stat --slow --time 5 # Find slowest requests in last 5 seconds
```

## Architecture

### Directory Structure

- `/bpf/` - eBPF kernel programs (C) and loader code
  - `*.bpf.c` - eBPF source code
  - `*.h` - Shared headers (data structures between kernel/userspace)
  - `loader/` - BPF program loading logic
  - Uses `go:generate` to compile eBPF programs
- `/agent/` - Main userspace agent
  - `protocol/` - Protocol parsers (each protocol in subdirectory)
  - `conn/` - Connection tracking and management
  - `analysis/` - Statistics and analysis engine
  - `render/` - Output rendering (watch TUI, stat aggregation)
  - `uprobe/` - Dynamic uprobe management for SSL/TLS
- `/cmd/` - CLI commands (Cobra framework)
- `/common/` - Shared utilities, logging, kernel version detection
- `/vmlinux/` - Pre-generated vmlinux.h headers per architecture

### Data Flow

1. eBPF programs hook syscalls (read/write) to capture raw protocol data
2. Data sent to userspace via perf ring buffer
3. Agent buffers data per connection, parses protocol messages
4. Matches request/response pairs, applies filters
5. Renders output (watch table or stat aggregation)

### Protocol Parsing System

Protocol parsers implement `ProtocolStreamParser` interface:
- `ParseStream()` - Parse messages from buffer
- `FindBoundary()` - Find message boundaries for recovery
- `Match()` - Match requests with responses

Parsers are registered in `ParsersMap` via `init()` functions.

## Adding a New Protocol

See `/docs/how-to-add-a-new-protocol.md` for detailed guide:

1. Create `/agent/protocol/<name>/` directory
2. Define message types (embed `FrameBase`)
3. Implement `ProtocolStreamParser` interface
4. Add kernel-space protocol inference in `/bpf/protocol_inference.h`
5. Add CLI subcommand in `/cmd/<name>.go`
6. Register parser in `ParsersMap`
7. Add e2e tests in `/testdata/`

## Key Dependencies

- `github.com/cilium/ebpf` - eBPF loading (pure Go, no libc)
- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/bubbletea` - TUI framework
- Go 1.23+, Clang 10+, LLVM 10+

## Logging

Multiple loggers available: `AgentLog`, `BPFLog`, `ProtocolParserLog`, etc.
Enable debug logging: `--debug --agent-log-level 5`
Protocol parsing debug: `--protocol-log-level 5`

## Testing

```bash
make test-go                    # Run all Go tests
go test -v ./agent/protocol/... # Test specific package
```

E2E tests are in `/testdata/` - shell scripts that run claudeinsight against test servers.

## Requirements

- Linux kernel 3.10 (from 3.10.0-957) or 4.14+
- Root privileges or CAP_BPF capability
- For kernels without BTF: use `--btf` flag with external BTF file
