//go:build ignore

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>

#define MAX_DATA_SIZE 4096

// SSL event structure - shared between kernel and user space
struct ssl_event {
    __u32 pid;
    __u32 data_len;
    __u8 is_read;  // 0=write, 1=read
    char comm[16];
    char data[MAX_DATA_SIZE];
} __attribute__((packed));

// RingBuffer Map for events
struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 256 * 1024);
} events SEC(".maps");

// PID filter map - 0 means monitor all processes
struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 1);
    __type(key, u32);
    __type(value, u32);
} target_pid_map SEC(".maps");

// Temporary map to store SSL_read arguments
struct ssl_read_args {
    __u64 buf;   // Store pointer as u64 for bpf2go compatibility
    __u64 num;
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 1024);
    __type(key, u64);  // pid_tgid
    __type(value, struct ssl_read_args);
} ssl_read_args_map SEC(".maps");

// Common function to capture SSL data
static __always_inline int capture_ssl_data(const void *buf, size_t num, u8 is_read) {
    // Filter invalid data
    if (num <= 0 || num > MAX_DATA_SIZE) {
        return 0;
    }

    // Mask to satisfy BPF verifier - ensures num is bounded
    u32 data_len = num & (MAX_DATA_SIZE - 1);
    if (data_len == 0 ) {
        return 0;
    }

    // PID filtering
    u64 pid_tgid = bpf_get_current_pid_tgid();
    u32 current_pid = pid_tgid >> 32;

    u32 key = 0;
    u32 *target_pid = bpf_map_lookup_elem(&target_pid_map, &key);
    if (target_pid && *target_pid != 0) {
        // If target PID is set, only capture that PID's data
        if (current_pid != *target_pid) {
            return 0;
        }
    }

    // Allocate event memory
    struct ssl_event *event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event) {
        return 0;
    }

    // Fill event data
    event->pid = current_pid;
    event->data_len = data_len;
    event->is_read = is_read;
    bpf_get_current_comm(&event->comm, sizeof(event->comm));

    // Copy plaintext data (data_len is verified bounded by & mask)
    if (bpf_probe_read_user(event->data, data_len, buf) != 0) {
        bpf_ringbuf_discard(event, 0);
        return 0;
    }

    // Submit event
    bpf_ringbuf_submit(event, 0);
    return 0;
}

// SSL_write hook - capture sent data
// Function prototype: int SSL_write(SSL *ssl, const void *buf, int num);
SEC("uprobe/SSL_write")
int BPF_UPROBE(uprobe_ssl_write, void *ssl, const void *buf, int num) {
    return capture_ssl_data(buf, num, 0);  // 0 = write
}

// SSL_read entry probe - save arguments
// Function prototype: int SSL_read(SSL *ssl, void *buf, int num);
SEC("uprobe/SSL_read")
int BPF_UPROBE(uprobe_ssl_read_entry, void *ssl, void *buf, int num) {
    u64 pid_tgid = bpf_get_current_pid_tgid();

    struct ssl_read_args args = {
        .buf = (__u64)buf,
        .num = (__u64)num,
    };

    bpf_map_update_elem(&ssl_read_args_map, &pid_tgid, &args, BPF_ANY);
    return 0;
}

// SSL_read return probe - capture actual data
SEC("uretprobe/SSL_read")
int BPF_URETPROBE(uprobe_ssl_read_exit, int ret) {
    u64 pid_tgid = bpf_get_current_pid_tgid();
    u32 current_pid = pid_tgid >> 32;

    // PID filtering
    u32 key = 0;
    u32 *target_pid = bpf_map_lookup_elem(&target_pid_map, &key);
    if (target_pid && *target_pid != 0) {
        if (current_pid != *target_pid) {
            goto cleanup;
        }
    }

    // Lookup saved arguments
    struct ssl_read_args *args = bpf_map_lookup_elem(&ssl_read_args_map, &pid_tgid);
    if (!args) {
        return 0;
    }

    // Check return value (actual bytes read)
    if (ret <= 0 || ret > MAX_DATA_SIZE) {
        goto cleanup;
    }

    // Ensure data_len is positive, use bitwise AND to satisfy BPF verifier
    u32 data_len = ret & (MAX_DATA_SIZE - 1);
    if (data_len == 0 || data_len > MAX_DATA_SIZE) {
        goto cleanup;
    }

    // Inline data capture logic (avoid function call for verifier)
    struct ssl_event *event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event) {
        goto cleanup;
    }

    event->pid = pid_tgid >> 32;
    event->data_len = data_len;
    event->is_read = 1;  // 1 = read
    bpf_get_current_comm(&event->comm, sizeof(event->comm));

    // Copy plaintext data
    if (bpf_probe_read_user(event->data, data_len, (void *)args->buf) != 0) {
        bpf_ringbuf_discard(event, 0);
        goto cleanup;
    }

    // Submit event
    bpf_ringbuf_submit(event, 0);

cleanup:
    bpf_map_delete_elem(&ssl_read_args_map, &pid_tgid);
    return 0;
}

char LICENSE[] SEC("license") = "GPL";
