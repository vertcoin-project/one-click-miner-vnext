package backend

import (
	"strings"
	"testing"

	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

func TestGPUCompatibility(t *testing.T) {
	intel := util.GPU{OSName: "Intel(R) UHD Graphics 630", Type: util.GPUTypeIntel}
	nvidia := util.GPU{OSName: "NVIDIA GeForce RTX 3060", Type: util.GPUTypeNVidia}
	basic := util.GPU{OSName: "Microsoft Basic Display Adapter", Type: util.GPUTypeOther}

	cases := []struct {
		name       string
		gpus       []util.GPU
		integrated bool
		wantErr    string // "" for no error
	}{
		{"dedicated GPU", []util.GPU{nvidia}, false, ""},
		{"dedicated and Intel", []util.GPU{intel, nvidia}, false, ""},
		{"Intel only, integrated off", []util.GPU{intel}, false, "Only integrated graphics found"},
		{"Intel only, integrated on", []util.GPU{intel}, true, ""},
		{"Intel and a basic adapter, integrated off", []util.GPU{basic, intel}, false, "Only integrated graphics found"},
		{"basic adapter only", []util.GPU{basic}, true, "No compatible GPUs detected"},
		{"nothing found", nil, false, "No compatible GPUs detected"},
	}
	for _, c := range cases {
		err := gpuCompatibility(c.gpus, "", c.integrated)
		switch {
		case c.wantErr == "" && err != nil:
			t.Errorf("%s: unexpected error %v", c.name, err)
		case c.wantErr != "" && (err == nil || !strings.HasPrefix(err.Error(), c.wantErr)):
			t.Errorf("%s: error %v, want %q", c.name, err, c.wantErr)
		}
	}
}
