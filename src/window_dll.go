//go:build windows

package src

import (
	"syscall"
	"unsafe"
)

// kernel32 is the handle to the Windows kernel DLL, shared across all
// Windows-specific files in this package.
var kernel32 = syscall.NewLazyDLL("kernel32.dll")

// Proc handles loaded once at startup.
var (
	procGetConsoleMode        = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode        = kernel32.NewProc("SetConsoleMode")
	procGetConsoleProcessList = kernel32.NewProc("GetConsoleProcessList")
)

// mustUnsafe is a convenience wrapper to pass a pointer as a uintptr.
func ptr(v interface{}) uintptr {
	return uintptr(unsafe.Pointer(&v))
}
