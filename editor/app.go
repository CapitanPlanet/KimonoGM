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
	"io"
	"time"
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
	if path == "" {
		path = a.projectPath
	}
	if path == "" {
		return fmt.Errorf("brak sciezki projektu")
	}
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
func (a *App) RunGame(projectPath string) error {
        // zabij poprzednie instancje na 5020
        exec.Command("taskkill", "/F", "/IM", "MakerEngine.exe").Run()
        exec.Command("cmd", "/C", "for /f \"tokens=5\" %a in ('netstat -aon ^| findstr :5020') do taskkill /F /PID %a").Run()

        enginePath, _ := filepath.Abs(filepath.Join("..", "engine", "MakerEngine.csproj"))
        if _, err := os.Stat(enginePath); err != nil {
                enginePath, _ = filepath.Abs(filepath.Join("..", "..", "engine", "MakerEngine.csproj"))
        }
        if _, err := os.Stat(enginePath); err != nil {
                return fmt.Errorf("nie znaleziono %s: %v", enginePath, err)
        }

        // forward slashes + cudzysłów dla ścieżek z spacją
        gp := filepath.ToSlash(projectPath)
        prop := fmt.Sprintf("-p:GameAssetsRoot=\"%s\"", gp)

        // to samo co Ci działa z konsoli + wyłączony trimming
        cmd := exec.Command("dotnet", "run", "--project", enginePath, prop, "-p:PublishTrimmed=false", "-p:BlazorWebAssemblyEnableTrimming=false", "--urls", "http://localhost:5020", "--no-launch-profile")
        cmd.Dir = filepath.Dir(enginePath)
        // log do pliku żebyś widział w edytorze co się wywala
        logFile, _ := os.Create(filepath.Join(os.TempDir(), "kimonogm_run.log"))
        if logFile != nil {
                cmd.Stdout = io.MultiWriter(os.Stdout, logFile)
                cmd.Stderr = io.MultiWriter(os.Stderr, logFile)
        } else {
                cmd.Stdout = os.Stdout
                cmd.Stderr = os.Stderr
        }
        err := cmd.Start()
        if err != nil {
                return err
        }
        // daj dotnetowi 2s na wstanie
        go func() {
                time.Sleep(2 * time.Second)
                // nie blokuj
                cmd.Process.Release()
        }()
        return nil
}

func (a *App) BuildStandaloneExe(projectPath string) (string, error) {
	if projectPath == "" {
		projectPath = a.projectPath
	}
	if projectPath == "" {
		return "", fmt.Errorf("brak sciezki projektu")
	}

	// Nazwa exe z folderu gry
	projectName := filepath.Base(projectPath)
	safeName := strings.ReplaceAll(projectName, " ", "_")
	
	// Pobrane - tam ma wylądować exe
	downloads := ""
	if home, err := os.UserHomeDir(); err == nil {
		downloads = filepath.Join(home, "Downloads")
	} else {
		downloads = os.TempDir()
	}
	_ = os.MkdirAll(downloads, 0755)
	outExe := filepath.Join(downloads, safeName+".exe")

	// Znajdź builder - tools/builder/main.go
	builderDir := ""
	candidates := []string{
		filepath.Join("..", "tools", "builder"),
		filepath.Join("..", "..", "tools", "builder"),
		filepath.Join("tools", "builder"),
		`C:\Users\MD-Core\KimonoGM\tools\builder`,
	}
	for _, c := range candidates {
		abs, _ := filepath.Abs(c)
		if _, err := os.Stat(abs); err == nil {
			builderDir = abs
			break
		}
	}
	if builderDir == "" {
		return "", fmt.Errorf("nie znaleziono tools/builder - stwórz folder tools/builder z main.go")
	}

	fmt.Printf("[EDITOR] Budowanie %s -> %s via %s\n", projectPath, outExe, builderDir)

	// go run . --game X --out Y
	cmd := exec.Command("go", "run", ".", "--game", projectPath, "--out", outExe)
	cmd.Dir = builderDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Pokaż logi w edytorze przez runtime.Log
	if a.ctx != nil {
		runtime.LogInfo(a.ctx, fmt.Sprintf("[BUILD] Start: %s -> %s", projectName, outExe))
	}

	if err := cmd.Run(); err != nil {
		if a.ctx != nil {
			runtime.LogError(a.ctx, fmt.Sprintf("[BUILD] FAILED: %v", err))
		}
		return "", err
	}

	if a.ctx != nil {
		runtime.LogInfo(a.ctx, fmt.Sprintf("[BUILD] GOTOWE: %s", outExe))
	}

	return outExe, nil
}

// Wersja uproszczona - jeśli nie chcesz budować player-template, tylko sam publish + zip
func (a *App) BuildGameToDownloads(projectPath string) (string, error) {
	if projectPath == "" {
		projectPath = a.projectPath
	}
	projectName := filepath.Base(projectPath)
	
	downloads := ""
	if home, err := os.UserHomeDir(); err == nil {
		downloads = filepath.Join(home, "Downloads")
	}
	outDir := filepath.Join(downloads, projectName+"_web")
	
	// Re-use istniejącego BuildGame ale z output do Pobranych
	enginePath, _ := filepath.Abs(filepath.Join("..", "engine", "MakerEngine.csproj"))
	if _, err := os.Stat(enginePath); err != nil {
		enginePath, _ = filepath.Abs(filepath.Join("..", "..", "engine", "MakerEngine.csproj"))
	}
	
	os.MkdirAll(outDir, 0755)
	cmd := exec.Command("dotnet", "publish", enginePath, fmt.Sprintf("-p:GameAssetsRoot=%s", projectPath), "-c", "Release", "-o", outDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return outDir, nil
}


func (a *App) BuildGame(projectPath string) error {
	enginePath, _ := filepath.Abs(filepath.Join("..", "engine", "MakerEngine.csproj"))
	if _, err := os.Stat(enginePath); err != nil {
		enginePath, _ = filepath.Abs(filepath.Join("..", "..", "engine", "MakerEngine.csproj"))
	}
	projectName := filepath.Base(projectPath)
	out, _ := filepath.Abs(filepath.Join(filepath.Dir(projectPath), "..", "dist", projectName))
	os.MkdirAll(out, 0755)
	cmd := exec.Command("dotnet", "publish", enginePath, fmt.Sprintf("-p:GameAssetsRoot=%s", projectPath), "-c", "Release", "-o", out)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
func (a *App) GetRecentProjects() []RecentProject {
	path := a.recentFilePath()
	b, err := os.ReadFile(path)
	if err != nil {
		return []RecentProject{}
	}
	var list []RecentProject
	if err := json.Unmarshal(b, &list); err != nil {
		return []RecentProject{}
	}
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

// PROJECT CREATION - FIXED: AUTO 1 BG + 1 AV (av_tutorial.webp priority)
func (a *App) CreateProject(basePath string, name string) error {
	basePath = filepath.Clean(basePath)
	for _, d := range []string{"Data", "images", filepath.Join("sounds", "sfx"), filepath.Join("sounds", "voice"), filepath.Join("sounds", "music")} {
		if err := os.MkdirAll(filepath.Join(basePath, d), 0755); err != nil {
			return err
		}
	}

	// 1. KOPIUJ Z frontend/src/assets -> images/
	srcAssets := "frontend/src/assets"
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

	// 2. WYBIERZ DEFAULTY - TWOJ av_tutorial.webp MA PRIORYTET
	defaultBG := ""
	defaultAV := ""

	imagesDir := filepath.Join(basePath, "images")
	if files, err := os.ReadDir(imagesDir); err == nil {
		// najpierw szukaj dokładnie av_tutorial i bg_tutorial
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			ln := strings.ToLower(f.Name())
			if ln == "av_tutorial.webp" || ln == "av_tutorial.png" || ln == "av_tutorial.jpg" {
				defaultAV = "images/" + f.Name()
			}
			if ln == "bg_tutorial.webp" || ln == "bg_tutorial.png" || ln == "bg_tutorial.jpg" {
				defaultBG = "images/" + f.Name()
			}
		}
		// potem pierwszy av_ / bg_
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			n := strings.ToLower(f.Name())
			if defaultAV == "" && strings.HasPrefix(n, "av_") {
				defaultAV = "images/" + f.Name()
			}
			if defaultBG == "" && strings.HasPrefix(n, "bg_") {
				defaultBG = "images/" + f.Name()
			}
		}
		// fallback - cokolwiek
		if defaultBG == "" {
			for _, f := range files {
				if !f.IsDir() {
					n := strings.ToLower(f.Name())
					if strings.HasSuffix(n, ".webp") || strings.HasSuffix(n, ".jpg") || strings.HasSuffix(n, ".png") {
						defaultBG = "images/" + f.Name()
						break
					}
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

	// 3. PROJECT.JANPROJ Z AUTOMATYCZNYM AVATAREM
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

	// 4. DAY1 Z TYM TŁEM
	day1Content := fmt.Sprintf(`[
  { "Id": "start", "SceneTitle": "Manieczki - Wejscie", "Background": "%s", "Text": "Wchodzisz do Manieczek. Kula dysko kreci sie nad glowa, a przy barze kolejka jak za komuny. Janusz w zlotym kimono juz tanczy.", "Choices": [{"Text": "Zamow oranzade przy barze", "Next": "bar"}, {"Text": "Wbij na parkiet", "Next": "parkiet"}], "Type": "normal", "Day": 1 },
  { "Id": "bar", "SceneTitle": "Kolejka przy barze", "Background": "%s", "Text": "Stoisz w kolejce. Barmanka w kolorowym dresie nalewa cos co wyglada jak oranzada ale pachnie jak bimber.", "Choices": [{"Text": "Pogadaj z barmanka", "Next": "koniec_dnia_1"}], "Type": "normal", "Day": 1 },
  { "Id": "parkiet", "SceneTitle": "Parkiet", "Background": "%s", "Text": "Parkiet pelen. Tancerz w zlotym kimono robi piruet i patrzy na Ciebie.", "Choices": [{"Text": "Zatancz", "Next": "koniec_dnia_1"}], "Type": "normal", "Day": 1 },
  { "Id": "koniec_dnia_1", "SceneTitle": "KONIEC DNIA 1", "Background": "%s", "Text": "Pierwsza noc w Manieczkach zaliczona.", "IsEndDay": true, "Type": "end_of_day", "Day": 1, "Choices": [] }
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