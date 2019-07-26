package prerequisites

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func checkAmdgpuDriverInstalled() error {
	if runtime.GOOS == "linux" {
		Info := exec.Command("lsmod")
		History, _ := Info.Output()
		lines := strings.Split(string(History), "\n")
		for _, l := range lines {
			if strings.Contains(l, "amdgpu") {
				return nil
			}
		}
		return fmt.Errorf("AMD GPU Driver is not installed. You need to install it in order to run the miner")
	}

	// If we don't know, assume OK
	return nil
}
