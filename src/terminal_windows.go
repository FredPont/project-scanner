//go:build windows

package src

import (
	"os"
	"syscall"
	"unsafe"
)

// EnableANSI enables Virtual Terminal Processing on Windows so that
// ANSI escape codes are rendered correctly in cmd and PowerShell.
func EnableANSI() {
	const enableVirtualTerminalProcessing = 0x0004
	handle := syscall.Handle(os.Stdout.Fd())
	var mode uint32
	procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	procSetConsoleMode.Call(uintptr(handle), uintptr(mode|enableVirtualTerminalProcessing))
}
