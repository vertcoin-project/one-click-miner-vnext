package prerequisites

import (
	"github.com/vertcoin-project/one-click-miner-vnext/logging"
)

func Install(name string, install chan bool) error {
	logging.Infof("Installing prerequisite [%s]\n", name)
	switch name {
	case "msvcrt2013":
		return installVCRT2013(install)
	case "amddriverlinux":
		return checkAmdgpuDriverInstalled()
	case "nvidiadriverlinux":
		return checkNvidiaDriverInstalled()
	default:
		logging.Warnf("Unknown prerequisite requested: %s", name)
	}

	return nil
}
