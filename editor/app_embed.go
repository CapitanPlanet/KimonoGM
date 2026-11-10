package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Krok: Bundle do 1 pliku.
// Edytor ma wkompilowane w srodku: Launcher.exe + engine-publish + icon.ico
// Przy pierwszym odpaleniu wypakowuje je do %LOCALAPPDATA%\KimonoGM\
// Potem juz ich nie wypakowuje, tylko uzywa tych z AppLocal.
// Dzieki temu dystrybucja to 1 plik KimonoGM.exe, bez bundle.zip obok.

//go:embed all:embedded
var embeddedBundle embed.FS

// embeddedVersion - jak zmienisz silnik / Launcher, podbij to, zeby wymusic ponowne wypakowanie
const embeddedVersion = "2" // v2: marker nie wystarczy, sprawdzamy tez czy Launcher.exe i engine-publish naprawde istnieja

func embeddedBaseDir() string {
	cache, err := os.UserCacheDir() // %LOCALAPPDATA%
	if err != nil || cache == "" {
		if home, herr := os.UserHomeDir(); herr == nil {
			cache = filepath.Join(home, "AppData", "Local")
		}
	}
	return filepath.Join(cache, "KimonoGM")
}

// ensureEmbeddedExtracted wypakowuje wkompilowany bundle do AppLocal, jak go jeszcze nie ma albo jest stara wersja.
// Zwraca folder AppLocal/KimonoGM, czyli tam gdzie lezy Launcher.exe i engine-publish.
func ensureEmbeddedExtracted() (string, error) {
	base := embeddedBaseDir()
	marker := filepath.Join(base, "_embedded_ver.txt")
	if b, err := os.ReadFile(marker); err == nil && strings.TrimSpace(string(b)) == embeddedVersion {
		// Marker OK, ale sprawdz czy naprawde jest Launcher.exe i engine-publish
		// Stary bug: lokalny build z pustym embedded/ zapisywal marker "1", a nic nie wypakowal
		// i potem CI build myslal ze juz jest, a nie bylo nic.
		if info, err := os.Stat(filepath.Join(base, "Launcher.exe")); err == nil && !info.IsDir() && info.Size() > 1000000 {
			if _, err := os.Stat(filepath.Join(base, "engine-publish")); err == nil {
				return base, nil
			}
		}
		// Marker jest, ale plikow nie ma - wypakuj jeszcze raz, nie ufaj markerowi
	}
	// Sprawdz czy w ogole mamy cos wkompilowanego
	entries, err := fs.ReadDir(embeddedBundle, "embedded")
	if err != nil || len(entries) == 0 {
		// Nie ma wkompilowanego bundle - to jest stary build z CI bez embedded, uzywamy plikow obok exe
		return "", fmt.Errorf("brak wkompilowanego bundle (embedded/)")
	}
	_ = os.MkdirAll(base, 0755)
	// Wypakuj wszystko z embedded/ do base/
	err = fs.WalkDir(embeddedBundle, "embedded", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel("embedded", p)
		if rel == "." || rel == "" {
			return nil
		}
		dst := filepath.Join(base, rel)
		if d.IsDir() {
			_ = os.MkdirAll(dst, 0755)
			return nil
		}
		b, err := embeddedBundle.ReadFile(p)
		if err != nil {
			return nil
		}
		_ = os.MkdirAll(filepath.Dir(dst), 0755)
		_ = os.WriteFile(dst, b, 0644)
		return nil
	})
	if err != nil {
		return "", err
	}
	_ = os.WriteFile(marker, []byte(embeddedVersion), 0644)
	return base, nil
}

// tryEmbeddedEngine szuka silnika w AppLocal po wypakowaniu embedded.
func tryEmbeddedEngine() string {
	base, err := ensureEmbeddedExtracted()
	if err != nil || base == "" {
		return ""
	}
	// base/engine-publish/wwwroot/index.html albo base/engine-publish/index.html
	for _, c := range []string{
		filepath.Join(base, "engine-publish"),
		filepath.Join(base, "engine-publish", "wwwroot"),
	} {
		if _, err := os.Stat(filepath.Join(c, "index.html")); err == nil {
			return filepath.Join(base, "engine-publish")
		}
	}
	// Jak nie ma index.html, ale jest folder, to i tak zwroc - findPublishedEngine sobie poradzi
	if _, err := os.Stat(filepath.Join(base, "engine-publish")); err == nil {
		return filepath.Join(base, "engine-publish")
	}
	return ""
}

// tryEmbeddedLauncher szuka Launcher.exe w AppLocal po wypakowaniu embedded.
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
