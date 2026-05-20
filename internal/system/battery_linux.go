//go:build linux

package system

// readBattery reads the first battery exposed under sysfs. Returns nil on
// desktops/servers with no battery.
func readBattery() *BatteryInfo {
	return readBatterySysfs("/sys/class/power_supply")
}
