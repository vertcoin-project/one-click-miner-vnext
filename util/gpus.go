package util

import (
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
	KnownGPU{"Radeon RX (Vega|[4-5][7-8]0)", GPUTypeAMD, nil},
	KnownGPU{"AMD Radeon VII", GPUTypeAMD, nil},
	KnownGPU{"NVIDIA GeForce GTX [0-9]{3,4}( ti)?", GPUTypeNVidia, nil},
	KnownGPU{"Advanced Micro Devices, Inc. \\[AMD/ATI\\] .*", GPUTypeAMD, nil},
	KnownGPU{"NVIDIA Corporation Device .*", GPUTypeNVidia, nil},
}

func init() {
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
