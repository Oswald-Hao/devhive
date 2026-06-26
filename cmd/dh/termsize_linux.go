//go:build !windows

package main

import (
	"syscall"
	"unsafe"
)

func getTermSizeFd(fd int) (int, int, error) {
	var ws struct{ Row, Col, Xpixel, Ypixel uint16 }
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if err != 0 {
		return 0, 0, err
	}
	return int(ws.Col), int(ws.Row), nil
}
