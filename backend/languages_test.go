package backend

import (
	"testing"

	"github.com/tidwall/buntdb"
)

func TestEveryLanguageHasAName(t *testing.T) {
	for _, l := range availableLanguages {
		if languageNames[l] == "" {
			t.Errorf("no name for language %s", l)
		}
	}
	if len(languageNames) != len(availableLanguages) {
		t.Errorf("%d names for %d languages", len(languageNames), len(availableLanguages))
	}
}

func TestLanguageWithoutSettings(t *testing.T) {
	m := &Backend{alreadyRunning: true}
	if got := m.GetLanguage(); got != "" {
		t.Errorf("GetLanguage() without settings = %q, want \"\"", got)
	}
}

func TestPickedLanguage(t *testing.T) {
	db, err := buntdb.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	m := &Backend{settings: db}

	if got := m.GetLanguage(); got != "" {
		t.Errorf("GetLanguage() before picking = %q, want \"\"", got)
	}

	m.SetLanguage("de")
	if got := m.GetLanguage(); got != "de" {
		t.Errorf("GetLanguage() after picking de = %q, want de", got)
	}
	if got := m.GetLocale(); got != "de" {
		t.Errorf("GetLocale() after picking de = %q, want de", got)
	}

	m.SetLanguage("xx")
	if got := m.GetLanguage(); got != "de" {
		t.Errorf("GetLanguage() after picking unknown xx = %q, want de", got)
	}

	m.SetLanguage("")
	if got := m.GetLanguage(); got != "" {
		t.Errorf("GetLanguage() after picking the system language = %q, want \"\"", got)
	}
}

func TestSystemLocaleIgnoresPickedLanguage(t *testing.T) {
	db, err := buntdb.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	m := &Backend{settings: db}

	system := m.GetSystemLocale()
	if !isAvailableLanguage(system) {
		t.Errorf("GetSystemLocale() = %q, not an available language", system)
	}
	picked := "de"
	if system == "de" {
		picked = "fr"
	}
	m.SetLanguage(picked)
	if got := m.GetSystemLocale(); got != system {
		t.Errorf("GetSystemLocale() after picking %s = %q, want %q", picked, got, system)
	}
}
