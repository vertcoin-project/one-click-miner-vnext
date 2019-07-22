package prerequisites

import (
	"fmt"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
)

func Install(name string, install chan bool) error {
	logging.Infof("Installing prerequisite [%s]\n", name)
	switch name {
	case "msvcrt2013":
		return installVCRT2013(install)
	default:
		return fmt.Errorf("Unknown prerequisite requested: %s", name)
	}
}
