.PHONY: all clean generate build run

# Variables
CLANG ?= clang
ARCH := $(shell uname -m | sed 's/x86_64/x86/' | sed 's/aarch64/arm64/')
BPF_OBJ := claude_monitor_bpfel.o
BINARY := claude_monitor

all: generate build

# Generate Go bindings from BPF C code
generate:
	@echo "Generating BPF bytecode and Go bindings..."
	go generate .

# Build the Go binary
build: generate
	@echo "Building Go binary..."
	go build -o $(BINARY) .

# Run the program (requires root)
run: build
	@echo "Running $(BINARY) (requires root privileges)..."
	sudo ./$(BINARY)

# Clean generated files
clean:
	@echo "Cleaning..."
	rm -f $(BINARY) $(BPF_OBJ) claude_monitor_bpfel.go claude_monitor_bpfeb.go

# Help
help:
	@echo "Available targets:"
	@echo "  all       - Generate BPF code and build (default)"
	@echo "  generate  - Generate vmlinux.h and Go bindings from BPF C code"
	@echo "  build     - Build the Go binary"
	@echo "  run       - Build and run the program (requires root)"
	@echo "  clean     - Remove generated files"
	@echo ""
	@echo "Usage examples:"
	@echo "  make              # Generate and build"
	@echo "  make run          # Build and run"
	@echo "  sudo ./$(BINARY) -pid 12345  # Monitor specific PID"
