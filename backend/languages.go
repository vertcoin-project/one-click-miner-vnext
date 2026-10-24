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
	"bg",
	"da",
	"de",
	"es",
	"fr",
	"hi",
	"hr",
	"it",
	"ja",
	"lt",
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

// languageNames are the names shown in the language picker, each written in
// its own language so it can be found without reading the current one
var languageNames = map[string]string{
	"en": "English",
	"bg": "Български",
	"da": "Dansk",
	"de": "Deutsch",
	"es": "Español",
	"fr": "Français",
	"hi": "हिन्दी",
	"hr": "Hrvatski",
	"it": "Italiano",
	"ja": "日本語",
	"lt": "Lietuvių",
	"nl": "Nederlands",
	"no": "Norsk",
	"pa": "ਪੰਜਾਬੀ",
	"pl": "Polski",
	"pt": "Português",
	"ro": "Română",
	"ru": "Русский",
	"sl": "Slovenščina",
	"sv": "Svenska",
	"tr": "Türkçe",
	"zh": "中文",
}

type LanguageChoice struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// GetLanguages returns the languages the user can pick in the settings
func (m *Backend) GetLanguages() []LanguageChoice {
	lc := make([]LanguageChoice, 0, len(availableLanguages))
	for _, l := range availableLanguages {
		lc = append(lc, LanguageChoice{Code: l, Name: languageNames[l]})
	}
	return lc
}

func isAvailableLanguage(l string) bool {
	for _, a := range availableLanguages {
		if a == l {
			return true
		}
	}
	return false
}

// GetLanguage returns the language picked in the settings, or "" when the
// system language is used
func (m *Backend) GetLanguage() string {
	l := m.getStringSetting("language")
	if !isAvailableLanguage(l) {
		return ""
	}
	return l
}

// SetLanguage saves the language to use. "" follows the system language.
func (m *Backend) SetLanguage(l string) {
	if l != "" && !isAvailableLanguage(l) {
		logging.Warnf("Ignoring unknown language %s", l)
		return
	}
	logging.Infof("Setting language to [%s]", l)
	m.setStringSetting("language", l)
}

// GetLocale returns the language picked in the settings, or the available
// language closest to the system's when none is picked
func (m *Backend) GetLocale() string {
	if l := m.GetLanguage(); l != "" {
		logging.Infof("Returning picked locale %s", l)
		return l
	}
	return systemLocale()
}

// GetSystemLocale returns the available language closest to the system's,
// whatever language is picked
func (m *Backend) GetSystemLocale() string {
	return systemLocale()
}

func systemLocale() string {
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
