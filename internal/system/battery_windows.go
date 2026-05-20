//go:build windows

package system

import (
	"syscall"
	"unsafe"
)

type systemPowerStatus struct {
	ACLineStatus        byte
	BatteryFlag         byte
	BatteryLifePercent  byte
	SystemStatusFlag    byte
	BatteryLifeTime     uint32
	BatteryFullLifeTime uint32
}

// readBattery queries the Windows power status. Returns nil when there is no
// system battery (e.g. desktops/workstations).
func readBattery() *BatteryInfo {
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("GetSystemPowerStatus")
	var s systemPowerStatus
	r, _, _ := proc.Call(uintptr(unsafe.Pointer(&s)))
	if r == 0 {
		return nil
	}
	const noBattery = 128 // BatteryFlag bit: no system battery
	if s.BatteryFlag&noBattery != 0 || s.BatteryLifePercent == 255 {
		return nil
	}
	return &BatteryInfo{
		Percent:  float64(s.BatteryLifePercent),
		Charging: s.ACLineStatus == 1,
	}
}
