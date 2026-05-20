//go:build !windows && !linux && !darwin

package system

// readBattery is a stub on platforms without a dedicated implementation.
func readBattery() *BatteryInfo { return nil }
