//go:build !windows

package src

// enableANSI is a no-op on non-Windows platforms.
func EnableANSI() {}
