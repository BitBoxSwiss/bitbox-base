package supervisor

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// unsetClearnetIDB unsets (0 - download blocks over Tor) the ibdClearnetRedisKey if set.
// The key can only be set back to 1 (download blocks over clearnet) via RPC.
func (s *Supervisor) unsetClearnetIDB() (err error) {
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
func (s *Supervisor) restartUnit(unit string) error {
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
func (s *Supervisor) setBBBConfigValue(argument string, value string) error {
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

func (s *Supervisor) checkBlockHeight(minHeight int) (err error) {
	blockHeight, err := s.prometheus.QueryFloat64("bitcoin_blocks")
	if blockHeight < float64(minHeight) {
		return fmt.Errorf("current block height (%d) is lower than the minimal block height (%d)", int(blockHeight), minHeight)
	}
	return nil
}

func (s *Supervisor) disableBaseIBDState() (err error) {

	// Before the IBD state of the Base is disabled the block height is sanity checked.
	// Disabling the ibd state too early results in c-lightning scanning all blocks, which
	// takes up to multiple days.
	const minBlockHeight int = 596000 // block mined on 9/22/2019
	err = s.checkBlockHeight(minBlockHeight)
	if err != nil {
		return fmt.Errorf("could not disable ibd state: %s", err.Error())
	}

	// unset clearnet IBD redis key when the IBD is finished
	err = s.unsetClearnetIDB()
	if err != nil {
		return fmt.Errorf("could not unset ClearnetIDB: %s", err.Error())
	}

	err = s.setBBBConfigValue("bitcoin_ibd", "false")
	if err != nil {
		return fmt.Errorf("could not execute a 'bbb-config.sh set': %s", err.Error())
	}

	return nil
}
