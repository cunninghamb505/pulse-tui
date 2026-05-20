//go:build !windows

package system

// readBattery is a stub on non-Windows platforms for now; macOS/Linux battery
// support can be added here later.
func readBattery() *BatteryInfo { return nil }
