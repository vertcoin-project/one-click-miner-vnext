package backend

import (
	"github.com/cloudfoundry/jibber_jabber"
	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"golang.org/x/text/language"
)

var availableLanguages = []string{
	// First language is default. So
	// alphabetical order except for
	// this one
	"en",
	"da",
	"de",
	"es",
	"fr",
	"hi",
	"hr",
	"it",
	"ja",
	"nl",
	"no",
	"pa",
	"pl",
	"pt",
	"ro",
	"ru",
	"sl",
	"sv",
	"tr",
	"zh",
}

var languageMatcher language.Matcher

func init() {
	tags := []language.Tag{}
	for _, l := range availableLanguages {
		t, err := language.Parse(l)
		if err == nil {
			tags = append(tags, t)
		}
	}

	languageMatcher = language.NewMatcher(tags)
}

func (m *Backend) GetLocale() string {
	userLanguage, err := jibber_jabber.DetectIETF()
	if err != nil {
		logging.Warnf("Could not determine locale, defaulting to English: %s", err.Error())
		return "en"
	}

	logging.Infof("User IETF is %s", userLanguage)
	userTag, err := language.Parse(userLanguage)
	if err != nil {
		logging.Warnf("Could not parse user IETF: %s", err.Error())
	}
	tag, _, _ := languageMatcher.Match(userTag)
	logging.Infof("Matched tag is %s", tag.String())
	base, _ := tag.Base()
	logging.Infof("Returning locale %s", base.String())
	return base.String()
}
