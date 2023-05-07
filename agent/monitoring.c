//go:build ignore

#define __TARGET_ARCH_x86
#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_tracing.h>

#define MAX_STR_LEN 128
#define MAX_ARR_CNT 64
#define TASK_COMM_LEN 16

char __license[] SEC("license") = "Dual MIT/GPL";

struct event {
	u32 pid;
	u32 ppid;
	u32 uid;
	char pp_name[TASK_COMM_LEN];
	char line[MAX_STR_LEN];
};

// A map for storing ringbuffer.
struct {
	__uint(type, BPF_MAP_TYPE_RINGBUF);
	__uint(max_entries, 1 << 24);
} events SEC(".maps");

// Force emitting struct event into the ELF.
const struct event *unused __attribute__((unused));

void det_strcpy(char *src, char *dst) {
    for (u32 i = 0 ; i < MAX_STR_LEN ; i++) {
        if (src[i] == 0) return;
        dst[i] = src[i];
    }
}

SEC("uretprobe/bash_readline")
int uretprobe_bash_readline(struct pt_regs *ctx) {
    // Get PID of this system call.
    u64 id   = bpf_get_current_pid_tgid();
    u32 tgid = id >> 32;

    // Get readline's input.
    char line[MAX_STR_LEN];
	bpf_probe_read(line, sizeof(line), (void *)PT_REGS_RC(ctx));

    // Get PPID of this process.
	struct task_struct *task = (struct task_struct *)bpf_get_current_task_btf();
	if (!task) {
	    return 0;
	}
	u32 ppid = (u32) task->real_parent->pid;

    // Get uid of this process.
    const struct cred *c = task->cred;
    u32 uid = (u32) c->uid.val;

    // Reserve ring buffer for sending it to Userspace
    struct event *task_info = {0};
	task_info = bpf_ringbuf_reserve(&events, sizeof(struct event), 0);
	if (!task_info) {
		return 0;
	}

    // Copy data to ringbuffer
	task_info->pid = tgid;
	task_info->ppid = ppid;
	task_info->uid = uid;
    det_strcpy(line, task_info->line);

	// Send ringbuffer data..
	bpf_ringbuf_submit(task_info, 0);

	return 0;
}