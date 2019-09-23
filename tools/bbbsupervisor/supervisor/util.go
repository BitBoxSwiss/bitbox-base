package supervisor

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// unsetClearnetIDB unsets (0 - download blocks over Tor) the ibdClearnetRedisKey if set.
// The key can only be set back to 1 (download blocks over clearnet) via RPC.
func (s Supervisor) unsetClearnetIDB() (err error) {
	const ibdClearnetRedisKey string = "bitcoind:ibd-clearnet"
	isIBDClearnet, err := s.redis.GetInt(ibdClearnetRedisKey)
	if err != nil {
		return fmt.Errorf("getting redis key %s failed: %v", ibdClearnetRedisKey, err)
	}
	if isIBDClearnet == 1 {
		log.Printf("IDB finished. Setting %s to %d.\n", ibdClearnetRedisKey, 0)
		err := s.setBBBConfigValue("bitcoin_ibd_clearnet", "false")
		if err != nil {
			return fmt.Errorf("disabling bitcoin_ibd_clearnet via BBB config script failed: %v", err)
		}
	}
	return
}

// restartUnit restarts a systemd unit
func (s Supervisor) restartUnit(unit string) error {
	args := []string{"restart", unit}
	cmd := exec.Command("/bin/systemctl", args...)
	cmdAsString := "systemctl " + strings.Join(args, " ")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command %s threw an error %v", cmdAsString, err)
	}
	log.Printf("restartUnit: command '%v' executed.\n", cmdAsString)
	return nil
}

// setBBBConfigValue calls `bbb-config.sh set <argument> <value>`
func (s Supervisor) setBBBConfigValue(argument string, value string) error {
	args := []string{"set", argument, value}
	executable := "/usr/local/sbin/bbb-config.sh"
	cmd := exec.Command(executable, args...)
	cmdAsString := executable + " " + strings.Join(args, " ")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command %s threw an error %v", cmdAsString, err)
	}
	log.Printf("setBBBConfigValue: command '%v' executed.\n", cmdAsString)
	return nil
}
