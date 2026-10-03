package main

import (
    "fmt"
    "io/fs"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    goruntime "runtime"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) SetProjectPath(path string) { a.projectPath = filepath.Clean(path) }
func (a *App) GetProjectPath() string     { return a.projectPath }
func (a *App) GetDefaultProjectPath() string {
    exe, _ := os.Executable()
    root := filepath.Join(filepath.Dir(exe), "..", "..", "..")
    games := filepath.Join(root, "games")
    if _, err := os.Stat(games); err == nil { return games }
    home, _ := os.UserHomeDir()
    return filepath.Join(home, "Documents", "JanuszProjects")
}
func (a *App) SelectFolder() (string, error) {
    return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "Wybierz folder projektu"})
}
func (a *App) OpenProjectFolder(path string) error {
    if path == "" { path = a.projectPath }
    if path == "" { return fmt.Errorf("brak ścieżki") }
    path = filepath.Clean(path)
    var cmd *exec.Cmd
    switch goruntime.GOOS {
    case "windows": cmd = exec.Command("explorer", path)
    case "darwin": cmd = exec.Command("open", path)
    default: cmd = exec.Command("xdg-open", path)
    }
    return cmd.Start()
}
func (a *App) CreateProject(basePath string, name string) error {
    basePath = filepath.Clean(basePath)
    for _, d := range []string{"Data", "images", filepath.Join("sounds", "sfx"), filepath.Join("sounds", "voice"), filepath.Join("sounds", "music")} {
        if err := os.MkdirAll(filepath.Join(basePath, d), 0755); err != nil { return err }
    }
    janprojPath := filepath.Join(basePath, "project.janproj")
    janprojContent := fmt.Sprintf("{\n  \"gameName\": \"%s\",\n  \"author\": \"\",\n  \"version\": \"1.0.0\",\n  \"engineVersion\": \"2.0.0\",\n  \"startDay\": \"day1\",\n  \"startScene\": \"start\",\n  \"statsSystem\": { \"stats\": [ {\"id\": \"CEBULA\", \"name\": \"CEBULA\", \"initial\": 0}, {\"id\": \"WSTYD\", \"name\": \"WSTYD\", \"initial\": 0}, {\"id\": \"PORTFEL\", \"name\": \"PORTFEL\", \"initial\": 0}, {\"id\": \"REPUTACJA\", \"name\": \"REPUTACJA\", \"initial\": 0} ] },\n  \"avatarSystem\": {\"default\": \"\", \"rules\": []}\n}", name)
    if err := os.WriteFile(janprojPath, []byte(janprojContent), 0644); err != nil { return err }
    day1Path := filepath.Join(basePath, "Data", "day1.json")
    day1Content := "[\n  { \"Id\": \"start\", \"SceneTitle\": \"Dzień 1 - Start\", \"Background\": \"images/bg_tutorial.jpg\", \"Text\": \"Janusz budzi się.\", \"Choices\": [{\"Text\": \"Dalej\", \"Next\": \"koniec_dnia_1\"}], \"Type\": \"normal\", \"Day\": 1 },\n  { \"Id\": \"koniec_dnia_1\", \"SceneTitle\": \"KONIEC DNIA 1\", \"Background\": \"images/bg_tutorial.jpg\", \"Text\": \"Koniec dnia 1.\", \"IsEndDay\": true, \"Type\": \"end_of_day\", \"Day\": 1, \"Choices\": [] }\n]"
    if err := os.WriteFile(day1Path, []byte(day1Content), 0644); err != nil { return err }
    manifestPath := filepath.Join(basePath, "Data", "_manifest.json")
    _ = os.WriteFile(manifestPath, []byte("{\n  \"days\": [\"day1\"],\n  \"startDay\": \"day1\",\n  \"startScene\": \"start\",\n  \"version\": \"2026-10-01T00:00:00.000Z\"\n}"), 0644)
    srcAssets := "frontend/src/assets"
    if _, err := os.Stat(srcAssets); err == nil {
        _ = filepath.Walk(srcAssets, func(p string, info fs.FileInfo, err error) error {
            if err != nil || info.IsDir() { return nil }
            low := strings.ToLower(info.Name())
            if !(strings.HasSuffix(low, ".webp") || strings.HasSuffix(low, ".png") || strings.HasSuffix(low, ".jpg") || strings.HasSuffix(low, ".jpeg")) { return nil }
            dst := filepath.Join(basePath, "images", info.Name())
            if _, err := os.Stat(dst); os.IsNotExist(err) {
                if b, err := os.ReadFile(p); err == nil { _ = os.WriteFile(dst, b, 0644) }
            }
            return nil
        })
    }
    a.projectPath = basePath
    _ = a.AddRecentProject(basePath)
    return nil
}
