// +build darwin

#include <errno.h>
#include <stdlib.h>
#include <stdio.h>
#include <strings.h>
#include <libproc.h>
#include <unistd.h>

// This is declared in process_darwin.go
extern void go_darwin_append_proc(pid_t, pid_t, char *);

// Loads the process table and calls the exported Go function to insert
// the data back into the Go space.
//
// This function is implemented in C because while it would technically
// be possible to do this all in Go, I didn't want to go spelunking through
// header files to get all the structures properly. It is much easier to just
// call it in C and be done with it.

void darwinProcesses() {

  uid_t euid = geteuid();

  int pid_buf_size = proc_listpids(PROC_ALL_PIDS, 0, NULL, 0);
  int pid_count = pid_buf_size / sizeof(pid_t);

  pid_t* pids = malloc(pid_buf_size);
  bzero(pids, pid_buf_size);

  proc_listpids(PROC_ALL_PIDS, 0, pids, pid_buf_size);
  char path_buffer[PROC_PIDPATHINFO_MAXSIZE];

  int ppid = 0;

  for (int i=0; i < pid_count; i++) {
    if (pids[i] == 0) break;

    if (euid == 0) {
      // You need root permission to get proc_bsdinfo from some processes.
      // When you call following function with normal user permission you will
      // receive 'operation not permitted' error and it will be terminated.
      struct proc_bsdinfo bsdinfo;
      proc_pidinfo(pids[i], PROC_PIDTBSDINFO, 0, &bsdinfo, sizeof(struct proc_bsdinfo));
      ppid = bsdinfo.pbi_ppid;
    }

    bzero(path_buffer, PROC_PIDPATHINFO_MAXSIZE);
    if (proc_pidpath(pids[i], path_buffer, sizeof(path_buffer)) > 0) {
      go_darwin_append_proc(pids[i], ppid, path_buffer);
    }
  }
  free(pids);
}
