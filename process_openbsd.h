// +build openbsd

#include <errno.h>
#include <stdio.h>
#include <kvm.h>
#include <limits.h>
#include <sys/param.h>
#include <sys/sysctl.h>

extern void go_openbsd_append_proc(pid_t, pid_t, char *);

static inline int openbsdProcesses() {
	int nentries = 0;
	int i = 0;
	char errbuf[_POSIX2_LINE_MAX];
	struct kinfo_proc *result = NULL;
	kvm_t *kernel = kvm_openfiles(NULL, NULL, NULL, KVM_NO_FILES, errbuf);
	if (!kernel) {
		return errno; 
	} 
	result = kvm_getprocs(kernel, KERN_PROC_ALL, 0, sizeof(struct kinfo_proc), &nentries);
	for (i = 0; i < nentries; i++) {
		struct kinfo_proc *single = &result[i]; 
		go_openbsd_append_proc(
			single->p_pid,
			single->p_ppid,
			single->p_comm);
	}

	return 0;
}
