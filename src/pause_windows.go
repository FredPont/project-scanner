//go:build windows

package src

import (
	"fmt"
	"os"
	"unsafe"
)

// WaitIfDoubleClicked pauses before exit when launched by double-click,
// so the console window stays open long enough for the user to read output.
func WaitIfDoubleClicked() {
	if isDoubleClicked() {
		fmt.Fprint(os.Stderr, "\nAppuyez sur Entrée pour quitter...")
		fmt.Scanln()
	}
}

// isDoubleClicked returns true when this process is the only one attached
// to its console — which happens when Explorer opens a new window for it.
// A shell (cmd, PowerShell) always appears in the list too, giving count >= 2.
func isDoubleClicked() bool {
	pids := make([]uint32, 2)
	ret, _, _ := procGetConsoleProcessList.Call(
		uintptr(unsafe.Pointer(&pids[0])),
		uintptr(len(pids)),
	)
	return ret == 1
}
