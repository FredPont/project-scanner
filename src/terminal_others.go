//go:build !windows

package src

// EnableANSI is a no-op on non-Windows platforms.
func EnableANSI() {}
