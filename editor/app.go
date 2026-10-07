package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	goruntime "runtime"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed frontend/templates
var templatesFS embed.FS

type App struct {
	ctx         context.Context
	projectPath string
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	runtime.LogInfo(a.ctx, "[APP] KimonoGM - Manieczki edition")
}

// PATHS
func (a *App) SetProjectPath(path string) {
	clean := filepath.Clean(path)
	clean = strings.ReplaceAll(clean, "/", "\\")
	a.projectPath = clean
}
func (a *App) GetProjectPath() string { return a.projectPath }
func (a *App) GetDefaultProjectPath() string {
	if configDir, err := os.UserConfigDir(); err == nil {
		p := filepath.Join(configDir, "KimonoGM", "games")
		_ = os.MkdirAll(p, 0755)
		return p
	}
	if home, err := os.UserHomeDir(); err == nil {
		p := filepath.Join(home, "Documents", "KimonoGM", "games")
		_ = os.MkdirAll(p, 0755)
		return p
	}
	exe, _ := os.Executable()
	root := filepath.Join(filepath.Dir(exe), "..", "..", "..")
	games := filepath.Join(root, "games")
	_ = os.MkdirAll(games, 0755)
	return games
}
func (a *App) OpenProjectFolder(path string) error {
	if path == "" { path = a.projectPath }
	if path == "" { return fmt.Errorf("brak sciezki projektu") }
	cleanPath := filepath.Clean(path)
	var cmd *exec.Cmd
	switch goruntime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", cleanPath)
	case "darwin":
		cmd = exec.Command("open", cleanPath)
	default:
		cmd = exec.Command("xdg-open", cleanPath)
	}
	return cmd.Start()
}
func (a *App) SelectFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Wybierz folder projektu", CanCreateDirectories: true,
	})
}

// HELPERS for assets.go, io.go, audio.go
func (a *App) cleanAssetPath(p string) string {
	if p == "" { return "" }
	clean := filepath.Clean(strings.TrimSpace(p))
	clean = strings.ReplaceAll(clean, "\\", "/")
	clean = strings.TrimPrefix(clean, "/")
	return clean
}
func (a *App) isInsideProject(absPath string) bool {
	if a.projectPath == "" { return false }
	absPath = filepath.Clean(absPath)
	proj := filepath.Clean(a.projectPath)
	rel, err := filepath.Rel(proj, absPath)
	if err != nil { return false }
	if rel == "." { return true }
	if strings.HasPrefix(rel, "..") { return false }
	return true
}
func (a *App) safeJoin(elems ...string) (string, error) {
	if len(elems) == 0 { return "", fmt.Errorf("pusta sciezka") }
	joined := filepath.Join(elems...)
	clean := filepath.Clean(joined)
	if a.projectPath != "" {
		abs := clean
		if !filepath.IsAbs(clean) {
			abs = filepath.Join(a.projectPath, clean)
			abs = filepath.Clean(abs)
		}
		if !a.isInsideProject(abs) {
			// pozwol jesli to plik w Data/images
			if !strings.HasPrefix(abs, filepath.Clean(a.projectPath)) {
				return "", fmt.Errorf("sciezka poza projektem: %s", clean)
			}
		}
		return abs, nil
	}
	return clean, nil
}
// RECENT
type RecentProject struct {
	Path string `json:"path"`
	Name string `json:"name"`
}
func (a *App) recentFilePath() string {
	if configDir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(configDir, "KimonoGM", "recent.json")
	}
	return filepath.Join(os.TempDir(), "kimonogm_recent.json")
}
func (a *App) GetRecentProjects() []RecentProject {
	path := a.recentFilePath()
	b, err := os.ReadFile(path)
	if err != nil { return []RecentProject{} }
	var list []RecentProject
	if err := json.Unmarshal(b, &list); err != nil { return []RecentProject{} }
	return list
}
func (a *App) AddRecentProject(projectPath string) error {
	projectPath = filepath.Clean(projectPath)
	name := filepath.Base(projectPath)
	list := a.GetRecentProjects()
	newList := []RecentProject{{Path: projectPath, Name: name}}
	for _, r := range list {
		if r.Path != projectPath && len(newList) < 10 {
			newList = append(newList, r)
		}
	}
	b, _ := json.MarshalIndent(newList, "", "  ")
	_ = os.MkdirAll(filepath.Dir(a.recentFilePath()), 0755)
	return os.WriteFile(a.recentFilePath(), b, 0644)
}
func (a *App) ClearRecentProjects() error { return os.Remove(a.recentFilePath()) }

// PROJECT CREATION
func (a *App) CreateProject(basePath string, name string) error {
	basePath = filepath.Clean(basePath)
	for _, d := range []string{"Data", "images", filepath.Join("sounds", "sfx"), filepath.Join("sounds", "voice"), filepath.Join("sounds", "music")} {
		if err := os.MkdirAll(filepath.Join(basePath, d), 0755); err != nil {
			return err
		}
	}
	janprojContent := fmt.Sprintf(`{
  "gameName": "%s",
  "author": "",
  "version": "1.0.0",
  "engineVersion": "2.0.0",
  "startDay": "day1",
  "startScene": "start",
  "statsSystem": { "stats": [ {"id": "CEBULA", "name": "CEBULA", "initial": 0}, {"id": "WSTYD", "name": "WSTYD", "initial": 0}, {"id": "PORTFEL", "name": "PORTFEL", "initial": 100}, {"id": "REPUTACJA", "name": "REPUTACJA", "initial": 0} ] },
  "avatarSystem": {"default": "", "rules": []}
}`, name)

	if err := os.WriteFile(filepath.Join(basePath, "project.janproj"), []byte(janprojContent), 0644); err != nil {
		return err
	}

	day1Content := `[
  { "Id": "start", "SceneTitle": "Manieczki - Wejscie", "Background": "images/bg_tutorial.webp", "Text": "Wchodzisz do Manieczek. Kula dysko kreci sie nad glowa, a przy barze kolejka jak za komuny. Janusz w zlotym kimono juz tanczy.", "Choices": [{"Text": "Zamow oranzade przy barze", "Next": "bar"}, {"Text": "Wbij na parkiet", "Next": "parkiet"}], "Type": "normal", "Day": 1 },
  { "Id": "bar", "SceneTitle": "Kolejka przy barze", "Background": "images/bg_tutorial.webp", "Text": "Stoisz w kolejce. Barmanka w kolorowym dresie nalewa cos co wyglada jak oranzada ale pachnie jak bimber.", "Choices": [{"Text": "Pogadaj z barmanka", "Next": "koniec_dnia_1"}], "Type": "normal", "Day": 1 },
  { "Id": "parkiet", "SceneTitle": "Parkiet", "Background": "images/bg_tutorial.webp", "Text": "Parkiet pelen. Tancerz w zlotym kimono robi piruet i patrzy na Ciebie.", "Choices": [{"Text": "Zatancz", "Next": "koniec_dnia_1"}], "Type": "normal", "Day": 1 },
  { "Id": "koniec_dnia_1", "SceneTitle": "KONIEC DNIA 1", "Background": "images/bg_tutorial.webp", "Text": "Pierwsza noc w Manieczkach zaliczona.", "IsEndDay": true, "Type": "end_of_day", "Day": 1, "Choices": [] }
]`
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

	srcAssets := "frontend/src/assets"
	if _, err := os.Stat(srcAssets); err == nil {
		_ = filepath.Walk(srcAssets, func(p string, info fs.FileInfo, err error) error {
			if err != nil || info.IsDir() { return nil }
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
	a.projectPath = basePath
	_ = a.AddRecentProject(basePath)
	return nil
}