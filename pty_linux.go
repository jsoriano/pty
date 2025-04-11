//go:build linux
// +build linux

package pty

import (
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

func open() (pty, tty *os.File, err error) {
	const devPtmx = "/dev/ptmx"

	fd, err := syscall.Open(devPtmx, os.O_RDWR, 0)
	if err != nil {
		return nil, nil, err
	}
	// In case of error after this point, make sure we close the ptmx fd.
	defer func() {
		if err != nil {
			_ = syscall.Close(fd) // Best effort.
		}
	}()

	sname, err := ptsname(fd)
	if err != nil {
		return nil, nil, err
	}

	if err := unlockpt(fd); err != nil {
		return nil, nil, err
	}

	tfd, err := syscall.Open(sname, os.O_RDWR|syscall.O_NOCTTY, 0) //nolint:gosec // Expected Open from a variable.
	if err != nil {
		return nil, nil, err
	}

	return os.NewFile(uintptr(fd), devPtmx), os.NewFile(uintptr(tfd), sname), nil
}

func ptsname(fd int) (string, error) {
	var n _C_uint
	err := ioctl(uintptr(fd), syscall.TIOCGPTN, uintptr(unsafe.Pointer(&n))) //nolint:gosec // Expected unsafe pointer for Syscall call.
	if err != nil {
		return "", err
	}
	return "/dev/pts/" + strconv.Itoa(int(n)), nil
}

func unlockpt(fd int) error {
	var u _C_int
	// use TIOCSPTLCK with a pointer to zero to clear the lock
	return ioctl(uintptr(fd), syscall.TIOCSPTLCK, uintptr(unsafe.Pointer(&u))) //nolint:gosec // Expected unsafe pointer for Syscall call.
}
