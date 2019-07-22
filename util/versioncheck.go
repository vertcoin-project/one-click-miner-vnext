package util

import (
	"fmt"
	"strconv"
	"strings"
)

type GithubRelease struct {
	URL        string `json:"html_url"`
	Tag        string `json:"tag_name"`
	Body       string `json:"body"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

var releases []GithubRelease

func init() {
	GetJson("https://api.github.com/repos/vertcoin-project/one-click-miner-vnext/releases", &releases)
}

func GetLatestRelease() (GithubRelease, error) {
	for _, r := range releases {
		if !r.Draft {
			return r, nil
		}
	}
	return GithubRelease{}, fmt.Errorf("No release found")
}

func VersionStringToNumeric(ver string) int64 {
	verNum := int64(0)
	// split off suffix
	suffix := ""
	suffixIdx := strings.Index(ver, "-")
	if suffixIdx > -1 {
		suffix = ver[suffixIdx+1:]
		ver = ver[:suffixIdx]

		// Chop off possible git commit hash and "-dirty"
		suffixIdx = strings.Index(suffix, "-")
		if suffixIdx > -1 {
			suffix = suffix[:suffixIdx]
		}
	}

	if len(suffix) > 0 {
		if strings.Contains(suffix, "alpha") {
			verNum += -999
			suffix = strings.ReplaceAll(suffix, "alpha", "")
		}
		if strings.Contains(suffix, "beta") {
			verNum += -899
			suffix = strings.ReplaceAll(suffix, "beta", "")
		}
		suffixVal, _ := strconv.Atoi(suffix)
		verNum += int64(suffixVal)
	}

	versionParts := strings.Split(ver, ".")
	multiplier := int64(100000000)
	for _, v := range versionParts {
		verVal, _ := strconv.Atoi(v)
		verNum += (int64(verVal) * multiplier)
		multiplier /= 100
	}

	return verNum
}
