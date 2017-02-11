// +build solaris

package ps

// #include <fcntl.h>
// #include <procfs.h>
//
// int read_psinfo(const char *path, psinfo_t *psi) {
//   int fh;
//   int retval = 0;
//   if ((fh = open(path, O_RDONLY)) >= 0) {
//     if (read(fh, psi, sizeof(psinfo_t)) == -1) {
//       retval = 1;
//     }
//     close(fh);
//     return retval;
//   }
//   return 2;
// }
import "C"

import "fmt"

func (p *UnixProcess) Refresh() error {
	var psinfo C.psinfo_t
	path := fmt.Sprintf("/proc/%d/psinfo", p.pid)
	_, err := C.read_psinfo(C.CString(path), &psinfo)
	if err != nil {
		return err
	}
	p.ppid = int(psinfo.pr_ppid)
	p.binary = C.GoString(&psinfo.pr_fname[0])
	return nil
}
