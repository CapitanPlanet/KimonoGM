package main

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Bundle do 1 pliku - wersja ZIP.
// Edytor ma wkompilowany tylko 1 plik: embedded/bundle.zip
// (Launcher.exe + engine-publish + icon.ico, spakowane przez CI)
// Przy pierwszym odpaleniu rozpakowuje go do %LOCALAPPDATA%\KimonoGM\
// Dzieki temu KimonoGM.exe jest ~80-90 MB, a nie ~200 MB.
// Dev lokalny: jak nie ma bundle.zip, to zrob pusty dummy bundle.zip w editor/embedded/
// zeby wails dev sie w ogole zbudowal - pusty zip nie zapisuje markera.

//go:embed embedded/bundle.zip
var embeddedBundleZip []byte

// embeddedVersion - podbij jak zmienisz silnik / Launcher, zeby wymusic ponowne wypakowanie
const embeddedVersion = "3"

func embeddedBaseDir() string {
	cache, err := os.UserCacheDir() // %LOCALAPPDATA%
	if err != nil || cache == "" {
		if home, herr := os.UserHomeDir(); herr == nil {
			cache = filepath.Join(home, "AppData", "Local")
		}
	}
	return filepath.Join(cache, "KimonoGM")
}

// ensureEmbeddedExtracted rozpakowuje wkompilowany bundle.zip do AppLocal, jak go jeszcze nie ma albo jest stara wersja.
func ensureEmbeddedExtracted() (string, error) {
	base := embeddedBaseDir()
	marker := filepath.Join(base, "_embedded_ver.txt")
	if b, err := os.ReadFile(marker); err == nil && strings.TrimSpace(string(b)) == embeddedVersion {
		// Marker OK, ale sprawdz czy naprawde jest Launcher.exe i engine-publish
		// Nie ufamy samemu markerowi - stary bug z pustym embedded/
		if info, err := os.Stat(filepath.Join(base, "Launcher.exe")); err == nil && !info.IsDir() && info.Size() > 1000000 {
			if _, err := os.Stat(filepath.Join(base, "engine-publish")); err == nil {
				return base, nil
			}
		}
	}
	if len(embeddedBundleZip) == 0 {
		return "", fmt.Errorf("brak wkompilowanego bundle (embedded/bundle.zip)")
	}
	_ = os.MkdirAll(base, 0755)
	zr, err := zip.NewReader(bytes.NewReader(embeddedBundleZip), int64(len(embeddedBundleZip)))
	if err != nil {
		return "", err
	}
	if len(zr.File) == 0 {
		// Pusty / dummy bundle.zip z deva lokalnego - nie pisz markera, nie udawaj ze wypakowales
		// Dev lokalny i tak uzyje zrodel z repo, a nie embeda
		return "", fmt.Errorf("pusty bundle.zip (dev lokalny, brak wkompilowanego bundle)")
	}
	for _, f := range zr.File {
		dst := filepath.Join(base, filepath.FromSlash(f.Name))
		if !strings.HasPrefix(dst, base+string(os.PathSeparator)) && dst != base {
			continue
		}
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(dst, 0755)
			continue
		}
		_ = os.MkdirAll(filepath.Dir(dst), 0755)
		rc, err := f.Open()
		if err != nil {
			continue
		}
		out, err := os.Create(dst)
		if err != nil {
			rc.Close()
			continue
		}
		_, _ = io.Copy(out, rc)
		_ = out.Close()
		_ = rc.Close()
	}
	_ = os.WriteFile(marker, []byte(embeddedVersion), 0644)
	return base, nil
}

// tryEmbeddedEngine szuka silnika w AppLocal po rozpakowaniu embedded.
func tryEmbeddedEngine() string {
	base, err := ensureEmbeddedExtracted()
	if err != nil || base == "" {
		return ""
	}
	for _, c := range []string{
		filepath.Join(base, "engine-publish"),
		filepath.Join(base, "engine-publish", "wwwroot"),
	} {
		if _, err := os.Stat(filepath.Join(c, "index.html")); err == nil {
			return filepath.Join(base, "engine-publish")
		}
	}
	if _, err := os.Stat(filepath.Join(base, "engine-publish")); err == nil {
		return filepath.Join(base, "engine-publish")
	}
	return ""
}

// tryEmbeddedLauncher szuka Launcher.exe w AppLocal po rozpakowaniu embedded.
func tryEmbeddedLauncher() string {
	base, err := ensureEmbeddedExtracted()
	if err != nil || base == "" {
		return ""
	}
	c := filepath.Join(base, "Launcher.exe")
	if info, err := os.Stat(c); err == nil && !info.IsDir() && info.Size() > 1000000 {
		return c
	}
	return ""
}
