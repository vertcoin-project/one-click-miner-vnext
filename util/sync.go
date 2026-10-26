package util

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/btcsuite/fastsha256"
)

// SyncZip makes unpackPath hold exactly the files of the zip archive. Files
// that already match are left alone, missing or changed ones are written, and
// files that aren't in the archive are removed. A matching file is never
// opened for writing, so a miner still running from it doesn't get in the way.
func SyncZip(archive, unpackPath string) error {
	r, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer r.Close()

	keep := map[string]bool{}
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		target, err := archiveTarget(unpackPath, f.Name)
		if err != nil {
			return err
		}
		keep[target] = true
		rc, err := f.Open()
		if err != nil {
			return err
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return err
		}
		if err := syncFile(target, data, f.Mode()); err != nil {
			return err
		}
	}
	return removeOthers(unpackPath, keep)
}

// SyncTar is SyncZip for a .tar.gz archive.
func SyncTar(archive, unpackPath string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(gzf)
	keep := map[string]bool{}
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		target, err := archiveTarget(unpackPath, header.Name)
		if err != nil {
			return err
		}
		keep[target] = true
		data, err := io.ReadAll(tarReader)
		if err != nil {
			return err
		}
		if err := syncFile(target, data, 0755); err != nil {
			return err
		}
	}
	return removeOthers(unpackPath, keep)
}

// archiveTarget is where an archive entry goes under unpackPath. It refuses
// entries whose names, like ../../x.exe, would land outside it (Zip Slip).
func archiveTarget(unpackPath, name string) (string, error) {
	target := filepath.Join(unpackPath, name)
	if !strings.HasPrefix(target, filepath.Clean(unpackPath)+string(os.PathSeparator)) {
		return "", fmt.Errorf("%s: illegal file path", target)
	}
	return target, nil
}

// syncFile writes data to target unless target already holds it.
func syncFile(target string, data []byte, mode os.FileMode) error {
	if sum, err := ShaSum(target); err == nil {
		want := fastsha256.Sum256(data)
		if bytes.Equal(sum, want[:]) {
			return nil
		}
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	return os.WriteFile(target, data, mode)
}

// removeOthers deletes the files under unpackPath that aren't in keep, so
// nothing can be planted next to the miner.
func removeOthers(unpackPath string, keep map[string]bool) error {
	return filepath.Walk(unpackPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || keep[path] {
			return nil
		}
		return os.Remove(path)
	})
}
