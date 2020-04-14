package util

import (
	"log"
	"testing"
)

func TestNvidia(t *testing.T) {
	gpus := []string{"NVIDIA GeForce 930MX", "NVIDIA GeForce GTX 1660 Ti", "NVIDIA GeForce RTX 2070", "NVIDIA P106-100", "NVIDIA Corporation GM107 [GeForce GTX 750 Ti] (rev a2)"}
	g := GetGPUsFromStrings(gpus)
	for i, gpu := range g {
		if gpu.Type != GPUTypeNVidia {
			log.Printf("Did not detect %s as NVIDIA!", gpus[i])
			t.Fail()
		}
	}
}

func TestAMD(t *testing.T) {
	gpus := []string{"Radeon (TM) RX 480 Graphics", "AMD Radeon(TM) R7 Graphics", "Radeon RX 480", "Radeon (TM) RX 560 Graphics"}
	g := GetGPUsFromStrings(gpus)
	for _, gpu := range g {
		if gpu.Type != GPUTypeAMD {
			t.Fail()
		}
	}
}

func TestInvalid(t *testing.T) {
	gpus := []string{"NVIDIA GeForce GTX 580"}
	g := GetGPUsFromStrings(gpus)
	for _, gpu := range g {
		if gpu.Type != GPUTypeOther {
			t.Fail()
		}
	}
}
