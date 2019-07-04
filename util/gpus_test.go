package util

import "testing"

func TestGeforceMXPUs(t *testing.T) {
	gpus := []string{"NVIDIA GeForce 930MX"}
	g := GetGPUsFromStrings(gpus)
	if g[0].Type != GPUTypeNVidia {
		t.Fail()
	}
}
