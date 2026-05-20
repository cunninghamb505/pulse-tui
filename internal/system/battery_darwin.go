//go:build darwin

package system

import "os/exec"

// readBattery parses `pmset -g batt`. Returns nil on Macs without a battery
// (e.g. Mac mini / Mac Studio), where pmset reports no percentage.
func readBattery() *BatteryInfo {
	out, err := exec.Command("pmset", "-g", "batt").Output()
	if err != nil {
		return nil
	}
	return parsePmset(string(out))
}
