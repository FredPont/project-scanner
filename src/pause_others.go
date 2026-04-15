//go:build !windows

package src

// WaitIfDoubleClicked is a no-op on non-Windows platforms.
func WaitIfDoubleClicked() {}
