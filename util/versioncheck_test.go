package util

import (
	"fmt"
	"testing"
)

func TestVersionStrings(t *testing.T) {
	testStrings := []string{"0.1-alpha1", "0.1.1", "0.2.1", "1.0-alpha1", "1.0", "0.1-alpha22-abe3f3b-dirty"}
	for _, s := range testStrings {
		fmt.Printf("%s = %d\n", s, VersionStringToNumeric(s))
	}
}
