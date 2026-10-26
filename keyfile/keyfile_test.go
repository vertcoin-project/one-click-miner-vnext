package keyfile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPrivateKeyWrongPassword(t *testing.T) {
	// Point the data directory at a temp folder, so the real keyfile is
	// never touched.
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	t.Setenv("HOME", dir)
	if !strings.HasPrefix(keyFile(), dir) {
		t.Fatalf("keyfile %s is not under the temp folder", keyFile())
	}
	if err := os.MkdirAll(filepath.Dir(keyFile()), 0700); err != nil {
		t.Fatal(err)
	}
	if err := CreateKeyFile("right"); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadPrivateKey("right"); err != nil {
		t.Errorf("right password: %v", err)
	}
	if _, err := LoadPrivateKey("wrong"); !errors.Is(err, ErrWrongPassword) {
		t.Errorf("wrong password: got %v, want ErrWrongPassword", err)
	}

	// An unreadable keyfile is not a wrong password
	if err := os.WriteFile(keyFile(), []byte("short"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPrivateKey("right"); err == nil || errors.Is(err, ErrWrongPassword) {
		t.Errorf("short keyfile: got %v, want an error other than ErrWrongPassword", err)
	}
}
