package util

import (
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
)

type GPUType int

const (
	GPUTypeOther  GPUType = 0
	GPUTypeAMD    GPUType = 1
	GPUTypeNVidia GPUType = 2
	GPUTypeIntel  GPUType = 3
)

type GPU struct {
	OSName string
	Type   GPUType
}

type KnownGPU struct {
	RegExPattern string
	Type         GPUType
	RegExp       *regexp.Regexp
}

var gpusCache []GPU
var gpusCached = false

var knownGPUs = []KnownGPU{
	/*KnownGPU{"Radeon( \\(TM\\))?( RX)? (Vega|[4-5][6-9]0)", GPUTypeAMD, nil},
	KnownGPU{"AMD Radeon\\(TM\\) R[79] Graphics", GPUTypeAMD, nil},
	KnownGPU{"AMD Radeon VII", GPUTypeAMD, nil},
	KnownGPU{"NVIDIA P[0-9]{3}-[0-9]{3}", GPUTypeNVidia, nil},
	KnownGPU{"NVIDIA GeForce (RTX )?(GTX )?(10|16|20|[7-9])[0-9]{2}( ti)?(MX)?", GPUTypeNVidia, nil},
	KnownGPU{"Advanced Micro Devices, Inc. \\[AMD/ATI\\] .*", GPUTypeAMD, nil},*/
	KnownGPU{".*NVIDIA.*", GPUTypeNVidia, nil},
	KnownGPU{".*AMD.*", GPUTypeAMD, nil},
	KnownGPU{".*Intel.*", GPUTypeIntel, nil},
	KnownGPU{".*Radeon.*", GPUTypeAMD, nil},
}

func init() {
	if os.Getenv("OCM_VIRTUALBOX") == "1" {
		knownGPUs = append(knownGPUs, KnownGPU{".*VirtualBox.*", GPUTypeIntel, nil})
	}

	for i := range knownGPUs {
		knownGPUs[i].RegExp, _ = regexp.Compile(knownGPUs[i].RegExPattern)
	}
}

func GetGPUsFromStrings(names []string) []GPU {
	gpus := []GPU{}
	for _, n := range names {
		found := false
		for _, k := range knownGPUs {
			if k.RegExp.Match([]byte(n)) {
				logging.Debugf("GPU [%s] matched regex [%s]\n", n, k.RegExp)
				gpus = append(gpus, GPU{n, k.Type})
				found = true
				break
			}
		}
		if !found {
			logging.Debugf("Unmatched GPU: [%s]\n", n)
			gpus = append(gpus, GPU{n, GPUTypeOther})
		}
	}
	return gpus
}

func GetGPUs() []GPU {
	if !gpusCached {
		gpus := []string{}
		if runtime.GOOS == "windows" {
			info := exec.Command("cmd", "/C", "wmic path win32_VideoController get name")
			PrepareBackgroundCommand(info)
			history, _ := info.Output()
			possibleGpus := strings.Split(string(history), "\n")
			for _, g := range possibleGpus {
				g = strings.Trim(g, "\r ")
				if g != "" && g != "Name" {
					gpus = append(gpus, g)
				}
			}
		} else if runtime.GOOS == "linux" {
			Info := exec.Command("lspci")
			History, _ := Info.Output()
			lines := strings.Split(string(History), "\n")
			for _, l := range lines {
				vgaIdx := strings.Index(l, "VGA compatible: ")
				if vgaIdx > -1 {
					gpus = append(gpus, l[vgaIdx+16:])
				}

				vgaIdx = strings.Index(l, "VGA compatible controller: ")
				if vgaIdx > -1 {
					gpus = append(gpus, l[vgaIdx+27:])
				}

				vgaIdx = strings.Index(l, "3D controller: ")
				if vgaIdx > -1 {
					gpus = append(gpus, l[vgaIdx+15:])
				}
			}
		} else if runtime.GOOS == "darwin" {
			Info := exec.Command("system_profiler", "SPDisplaysDataType")
			History, _ := Info.Output()
			lines := strings.Split(string(History), "\n")
			for _, l := range lines {
				csIdx := strings.Index(l, "Chipset Model: ")
				if csIdx > -1 {
					gpus = append(gpus, l[csIdx+15:])
				}
			}
		}
		gpusCache = GetGPUsFromStrings(gpus)
		gpusCached = true
	}
	return gpusCache
}
