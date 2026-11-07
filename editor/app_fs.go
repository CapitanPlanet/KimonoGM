package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (a *App) cleanAssetPath(p string) string {
	if p == "" {
		return ""
	}
	clean := filepath.Clean(strings.TrimSpace(p))
	clean = strings.ReplaceAll(clean, "\\", "/")
	clean = strings.TrimPrefix(clean, "/")
	return clean
}

func (a *App) isInsideProject(absPath string) bool {
	if a.projectPath == "" {
		return false
	}
	absPath = filepath.Clean(absPath)
	proj := filepath.Clean(a.projectPath)
	rel, err := filepath.Rel(proj, absPath)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	if strings.HasPrefix(rel, "..") {
		return false
	}
	return true
}

func (a *App) safeJoin(elems ...string) (string, error) {
	if len(elems) == 0 {
		return "", fmt.Errorf("pusta sciezka")
	}
	joined := filepath.Join(elems...)
	clean := filepath.Clean(joined)
	if a.projectPath != "" {
		abs := clean
		if !filepath.IsAbs(clean) {
			abs = filepath.Join(a.projectPath, clean)
			abs = filepath.Clean(abs)
		}
		if !a.isInsideProject(abs) {
			if !strings.HasPrefix(abs, filepath.Clean(a.projectPath)) {
				return "", fmt.Errorf("sciezka poza projektem: %s", clean)
			}
		}
		return abs, nil
	}
	return clean, nil
}

func (a *App) findEngineRoot() (string, string) {
	// Inteligentna funkcja: wspina się w górę katalogów w poszukiwaniu silnika
	checkDir := func(startDir string) (string, string) {
		current := startDir
		for i := 0; i < 8; i++ { // sprawdź do 8 poziomów w górę
			engineCandidate := filepath.Join(current, "engine")
			if _, err := os.Stat(filepath.Join(engineCandidate, "MakerEngine.csproj")); err == nil {
				return engineCandidate, filepath.Join(filepath.Dir(engineCandidate), "dist")
			}
			if _, err := os.Stat(filepath.Join(current, "MakerEngine.csproj")); err == nil {
				if filepath.Base(current) == "engine" {
					return current, filepath.Join(filepath.Dir(current), "dist")
				}
			}
			parent := filepath.Dir(current)
			if parent == current {
				break
			}
			current = parent
		}
		return "", ""
	}

	// 1. Szukaj w górę od pliku wykonywalnego (.exe)
	if exe, err := os.Executable(); err == nil {
		if abs, dist := checkDir(filepath.Dir(exe)); abs != "" {
			return abs, dist
		}
	}

	// 2. Szukaj w górę od bieżącego katalogu roboczego (CWD)
	if cwd, err := os.Getwd(); err == nil {
		if abs, dist := checkDir(cwd); abs != "" {
			return abs, dist
		}
	}

	// 3. Awaryjne ścieżki stałe na wypadek sandboksa
	candidates := []string{
		`C:\Users\MD-Core\KimonoGM\engine`,
		`C:\KimonoGM\engine`,
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(c, "MakerEngine.csproj")); err == nil {
			return c, filepath.Join(filepath.Dir(c), "dist")
		}
	}

	return "", ""
}

func getDownloadsFolder() string {
	home := os.Getenv("USERPROFILE")
	for _, c := range []string{
		filepath.Join(home, "Downloads"),
		filepath.Join(home, "Pobrane"),
	} {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return os.TempDir()
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	os.MkdirAll(filepath.Dir(dst), 0755)
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// ---------- RECENT ----------
type RecentProject struct {
	Path string `json:"path"`
	Name string `json:"name"`
}