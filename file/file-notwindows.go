// +build !windows

package file

import (
	"os"
	"syscall"
)

func init() {
	lockFile = func(f *os.File) {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_SH)
	}
	unlockFile = func(f *os.File) {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	}
}
