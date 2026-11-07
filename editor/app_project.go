package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func (a *App) CreateProject(basePath string, name string) error {
	basePath = filepath.Clean(basePath)
	for _, d := range []string{"Data", "images", filepath.Join("sounds", "sfx"), filepath.Join("sounds", "voice"), filepath.Join("sounds", "music")} {
		if err := os.MkdirAll(filepath.Join(basePath, d), 0755); err != nil {
			return err
		}
	}

	imagesDir := filepath.Join(basePath, "images")

	// Bezpieczne namierzanie assetów źródłowych (działa w dev oraz w skompilowanym EXE)
	srcAssets := "frontend/src/assets"
	if _, err := os.Stat(srcAssets); os.IsNotExist(err) {
		if exe, err := os.Executable(); err == nil {
			candidate := filepath.Join(filepath.Dir(exe), "..", "..", "frontend", "src", "assets")
			if _, err := os.Stat(candidate); err == nil {
				srcAssets = candidate
			} else {
				candidate2 := filepath.Join(filepath.Dir(exe), "..", "frontend", "src", "assets")
				if _, err := os.Stat(candidate2); err == nil {
					srcAssets = candidate2
				}
			}
		}
	}

	// Kopiujemy prawdziwe aktywa ze ścieżki deweloperskiej
	if _, err := os.Stat(srcAssets); err == nil {
		_ = filepath.Walk(srcAssets, func(p string, info fs.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			low := strings.ToLower(info.Name())
			if !(strings.HasSuffix(low, ".webp") || strings.HasSuffix(low, ".png") || strings.HasSuffix(low, ".jpg") || strings.HasSuffix(low, ".jpeg")) {
				return nil
			}
			dst := filepath.Join(basePath, "images", info.Name())
			if _, err := os.Stat(dst); os.IsNotExist(err) {
				if b, err := os.ReadFile(p); err == nil {
					_ = os.WriteFile(dst, b, 0644)
				}
			}
			return nil
		})
	}

	// Inteligentne wykrywanie domyślnego tła i awatara
	defaultBG := ""
	defaultAV := ""
	if files, err := os.ReadDir(imagesDir); err == nil {
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			ln := strings.ToLower(f.Name())
			if defaultAV == "" && (strings.HasPrefix(ln, "av_") || strings.HasPrefix(ln, "avatar")) {
				defaultAV = "images/" + f.Name()
			}
			if defaultBG == "" && (strings.HasPrefix(ln, "bg_") || strings.HasPrefix(ln, "background")) {
				defaultBG = "images/" + f.Name()
			}
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			n := strings.ToLower(f.Name())
			if strings.HasSuffix(n, ".webp") || strings.HasSuffix(n, ".png") || strings.HasSuffix(n, ".jpg") {
				if defaultAV == "" {
					defaultAV = "images/" + f.Name()
				}
				if defaultBG == "" {
					defaultBG = "images/" + f.Name()
				}
			}
		}
	}

	if defaultBG == "" {
		defaultBG = "images/bg_tutorial.webp"
	}
	if defaultAV == "" {
		defaultAV = "images/av_tutorial.webp"
	}

	janprojContent := fmt.Sprintf(`{
  "gameName": "%s",
  "author": "",
  "version": "1.0.0",
  "engineVersion": "2.0.0",
  "startDay": "day1",
  "startScene": "start",
  "statsSystem": { "stats": [ {"id": "CEBULA", "name": "CEBULA", "initial": 0}, {"id": "WSTYD", "name": "WSTYD", "initial": 0}, {"id": "PORTFEL", "name": "PORTFEL", "initial": 100}, {"id": "REPUTACJA", "name": "REPUTACJA", "initial": 0} ] },
  "avatarSystem": {"default": "%s", "rules": []}
}`, name, defaultAV)
	if err := os.WriteFile(filepath.Join(basePath, "project.janproj"), []byte(janprojContent), 0644); err != nil {
		return err
	}

	day1Content := fmt.Sprintf(`[
  { "Id": "start", "SceneTitle": "Manieczki - Wejscie", "Background": "%s", "Text": "Wchodzisz do Manieczek.", "Choices": [{"Text": "Zamow oranzade przy barze", "Next": "bar"}, {"Text": "Wbij na parkiet", "Next": "parkiet"}], "Type": "normal", "Day": 1 },
  { "Id": "bar", "SceneTitle": "Kolejka przy barze", "Background": "%s", "Text": "Stoisz w kolejce.", "Choices": [{"Text": "Pogadaj z barmanka", "Next": "koniec_dnia_1"}], "Type": "normal", "Day": 1 },
  { "Id": "parkiet", "SceneTitle": "Parkiet", "Background": "%s", "Text": "Parkiet pelen.", "Choices": [{"Text": "Zatancz", "Next": "koniec_dnia_1"}], "Type": "normal", "Day": 1 },
  { "Id": "koniec_dnia_1", "SceneTitle": "KONIEC DNIA 1", "Background": "%s", "Text": "Pierwsza noc zaliczona.", "IsEndDay": true, "Type": "end_of_day", "Day": 1, "Choices": [] }
]`, defaultBG, defaultBG, defaultBG, defaultBG)
	if err := os.WriteFile(filepath.Join(basePath, "Data", "day1.json"), []byte(day1Content), 0644); err != nil {
		return err
	}

	manifest := `{
  "days": ["day1"],
  "startDay": "day1",
  "startScene": "start",
  "version": "2026-10-04T00:00:00.000Z"
}`
	_ = os.WriteFile(filepath.Join(basePath, "Data", "_manifest.json"), []byte(manifest), 0644)

	a.projectPath = basePath
	_ = a.AddRecentProject(basePath)
	return nil
}