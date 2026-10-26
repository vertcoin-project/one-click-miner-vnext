package util

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var archiveFiles = map[string]string{
	"VerthashMiner/VerthashMiner.exe": "miner binary",
	"VerthashMiner/kernels/a.cl":      "kernel source",
}

func writeZip(t *testing.T, path string) {
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := zip.NewWriter(f)
	if _, err := w.Create("VerthashMiner/"); err != nil {
		t.Fatal(err)
	}
	for name, body := range archiveFiles {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

func writeTar(t *testing.T, path string) {
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for name, body := range archiveFiles {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0755, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSyncArchive(t *testing.T) {
	formats := []struct {
		name  string
		write func(*testing.T, string)
		sync  func(string, string) error
	}{
		{"zip", writeZip, SyncZip},
		{"tar.gz", writeTar, SyncTar},
	}
	for _, f := range formats {
		t.Run(f.name, func(t *testing.T) {
			dir := t.TempDir()
			archive := filepath.Join(dir, "archive")
			unpacked := filepath.Join(dir, "unpacked")
			exe := filepath.Join(unpacked, "VerthashMiner", "VerthashMiner.exe")
			kernel := filepath.Join(unpacked, "VerthashMiner", "kernels", "a.cl")
			extra := filepath.Join(unpacked, "VerthashMiner", "OpenCL.dll")
			f.write(t, archive)

			sync := func() {
				t.Helper()
				if err := f.sync(archive, unpacked); err != nil {
					t.Fatal(err)
				}
				for name, body := range archiveFiles {
					got, err := os.ReadFile(filepath.Join(unpacked, filepath.FromSlash(name)))
					if err != nil || string(got) != body {
						t.Errorf("%s = %q, %v; want %q", name, got, err, body)
					}
				}
			}

			// Into an empty folder
			sync()

			// A matching file is left alone
			old := time.Now().Add(-time.Hour).Truncate(time.Second)
			if err := os.Chtimes(exe, old, old); err != nil {
				t.Fatal(err)
			}
			sync()
			if info, _ := os.Stat(exe); !info.ModTime().Equal(old) {
				t.Errorf("matching %s was rewritten", exe)
			}

			// A changed file and a missing one are restored, a planted one removed
			if err := os.WriteFile(exe, []byte("tampered binary"), 0644); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(kernel); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(extra, []byte("planted"), 0644); err != nil {
				t.Fatal(err)
			}
			sync()
			if _, err := os.Stat(extra); !os.IsNotExist(err) {
				t.Errorf("planted %s was not removed: %v", extra, err)
			}
		})
	}
}
