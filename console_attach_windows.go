//go:build windows

package main

import (
	"os"
	"syscall"
)

const attachParentProcess = ^uintptr(0)

var (
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")
	procAttach  = modkernel32.NewProc("AttachConsole")
	procFree    = modkernel32.NewProc("FreeConsole")
)

func attachParentConsole() {
	r, _, _ := procAttach.Call(attachParentProcess)
	if r == 0 {
		return
	}
	redirect("CONOUT$", syscall.GENERIC_WRITE, &os.Stdout)
	redirect("CONOUT$", syscall.GENERIC_WRITE, &os.Stderr)
	redirect("CONIN$", syscall.GENERIC_READ, &os.Stdin)
}

func redirect(name string, access uint32, cur **os.File) {
	h, err := syscall.CreateFile(
		syscall.StringToUTF16Ptr(name),
		access,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
		nil,
		syscall.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil || h == syscall.InvalidHandle {
		return
	}
	f := os.NewFile(uintptr(h), name)
	if f == nil {
		return
	}
	*cur = f
}
