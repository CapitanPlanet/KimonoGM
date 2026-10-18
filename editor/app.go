
package main

import (
        "archive/zip"
        "context"
        "embed"
        "encoding/json"
        "fmt"
        "io"
        "io/fs"
        "net/http"
        "os"
        "os/exec"
        "path/filepath"
        "strings"
        "sync"
        "time"
        goruntime "runtime"

        "github.com/wailsapp/wails/v2/pkg/runtime"
)

var placeholderPNG = []byte{137, 80, 78, 71, 13, 10, 26, 10, 0, 0, 0, 13, 73, 72, 68, 82, 0, 0, 0, 1, 0, 0, 0, 1, 8, 6, 0, 0, 0, 31, 21, 196, 137, 0, 0, 0, 13, 73, 68, 65, 84, 120, 156, 98, 0, 1, 0, 0, 5, 0, 1, 13, 10, 45, 180, 0, 0, 0, 0, 73, 69, 78, 68, 174, 66, 96, 130}

//go:embed frontend/templates
var templatesFS embed.FS

type App struct {
        ctx         context.Context
        projectPath string
        hostServer  *http.Server
        hostMu      sync.Mutex
        hostPort    int
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) {
        a.ctx = ctx
        runtime.LogInfo(a.ctx, "[APP] KimonoGM - Manieczki edition")
}
func (a *App) shutdown(ctx context.Context) { exec.Command("taskkill", "/F", "/IM", "dotnet.exe").Run() }

// ---------- PATHS ----------
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

// ---------- HELPERS for assets.go, io.go, audio.go - FIX dla wails ----------
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
        cwd, _ := os.Getwd()
        candidates := []string{
                filepath.Join(filepath.Dir(filepath.Dir(cwd)), "engine"),
                filepath.Join(cwd, "..", "..", "engine"),
                `C:\Users\MD-Core\KimonoGM\engine`,
                filepath.Join(cwd, "..", "engine"),
                "engine",
                filepath.Join(filepath.Dir(cwd), "engine"),
        }
        for _, c := range candidates {
                abs, _ := filepath.Abs(c)
                if _, err := os.Stat(filepath.Join(abs, "MakerEngine.csproj")); err == nil {
                        return abs, filepath.Join(filepath.Dir(abs), "dist")
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
func (a *App) recentFilePath() string {
        if configDir, err := os.UserConfigDir(); err == nil {
                return filepath.Join(configDir, "KimonoGM", "recent.json")
        }
        return filepath.Join(os.TempDir(), "kimonogm_recent.json")
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
        var valid []RecentProject
        for _, r := range list {
                if _, err := os.Stat(r.Path); err == nil {
                        valid = append(valid, r)
                }
        }
        return valid
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

// ---------- SAVE META ----------
func (a *App) SaveProjectMeta(projectPath string, metaJson string) error {
        if projectPath == "" {
                projectPath = a.projectPath
        }
        var data interface{}
        if err := json.Unmarshal([]byte(metaJson), &data); err != nil {
                return err
        }
        outPath := filepath.Join(projectPath, "project.janproj")
        b, _ := json.MarshalIndent(data, "", "  ")
        return os.WriteFile(outPath, b, 0644)
}

// ---------- CREATE PROJECT ----------
func (a *App) CreateProject(basePath string, name string) error {
        basePath = filepath.Clean(basePath)
        for _, d := range []string{"Data", "images", filepath.Join("sounds", "sfx"), filepath.Join("sounds", "voice"), filepath.Join("sounds", "music")} {
                if err := os.MkdirAll(filepath.Join(basePath, d), 0755); err != nil {
                        return err
                }
        }
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
        defaultBG := ""
        defaultAV := ""
        imagesDir := filepath.Join(basePath, "images")
        if files, err := os.ReadDir(imagesDir); err == nil {
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


func (a *App) RunGame(projectPath string) error {
        if projectPath == "" {
                projectPath = a.projectPath
        }
        if projectPath == "" {
                return fmt.Errorf("brak sciezki projektu")
        }
        a.projectPath = filepath.Clean(projectPath)

        // 1. KILL
        exec.Command("taskkill", "/F", "/IM", "MakerEngine.exe").Run()
        exec.Command("taskkill", "/F", "/IM", "dotnet.exe").Run()
        exec.Command("cmd", "/C", "for /f \"tokens=5\" %a in ('netstat -aon ^| findstr :5020') do taskkill /F /PID %a").Run()
        time.Sleep(500 * time.Millisecond)

        // 2. FIND ENGINE
        engineRoot, _ := a.findEngineRoot()
        if engineRoot == "" {
                return fmt.Errorf("nie znaleziono engine/MakerEngine.csproj")
        }
        enginePath := filepath.Join(engineRoot, "MakerEngine.csproj")
        srcData := filepath.Join(projectPath, "Data")
        if _, err := os.Stat(srcData); os.IsNotExist(err) {
                srcData = projectPath
        }
        runtime.LogInfo(a.ctx, "[PARKIET] Project: "+projectPath+" -> Data: "+srcData)

        // 3. PRZYGOTUJ wwwroot
        _ = os.RemoveAll(filepath.Join(engineRoot, "obj"))
        _ = os.RemoveAll(filepath.Join(engineRoot, "bin"))
        _ = os.MkdirAll(filepath.Join(engineRoot, "wwwroot", "data", "images"), 0755)
        _ = os.MkdirAll(filepath.Join(engineRoot, "wwwroot", "images"), 0755)
        _ = os.MkdirAll(filepath.Join(engineRoot, "wwwroot", "data"), 0755)
        _ = os.MkdirAll(filepath.Join(engineRoot, "Assets", "data", "images"), 0755)
        _ = os.MkdirAll(filepath.Join(engineRoot, "Assets", "wwwroot", "data", "images"), 0755)

        // av_tutorial.png MUSI istniec WSZEDZIE przed buildem bo Assets->wwwroot kopiuje i kasuje
        allAvPaths := []string{
                filepath.Join(engineRoot, "wwwroot", "images", "av_tutorial.png"),
                filepath.Join(engineRoot, "wwwroot", "data", "images", "av_tutorial.png"),
                filepath.Join(engineRoot, "Assets", "data", "images", "av_tutorial.png"),
                filepath.Join(engineRoot, "Assets", "wwwroot", "data", "images", "av_tutorial.png"),
                filepath.Join(engineRoot, "Assets", "images", "av_tutorial.png"),
        }
        for _, p := range allAvPaths {
                _ = os.MkdirAll(filepath.Dir(p), 0755)
                if info, err := os.Stat(p); err != nil || info.Size() == 0 {
                        _ = os.WriteFile(p, placeholderPNG, 0644)
                        runtime.LogInfo(a.ctx, "[PARKIET] Utworzono placeholder "+p)
                }
        }

        // Funkcja kopiujaca do wszystkich mozliwych lokalizacji
        copyData := func() int {
                var c int
                targetDirs := []string{
                        filepath.Join(engineRoot, "wwwroot", "data"),
                        filepath.Join(engineRoot, "Assets", "data"),
                        filepath.Join(engineRoot, "Assets", "wwwroot", "data"),
                }
                targetImageDirs := []string{
                        filepath.Join(engineRoot, "wwwroot", "images"),
                        filepath.Join(engineRoot, "wwwroot", "data", "images"),
                        filepath.Join(engineRoot, "Assets", "data", "images"),
                        filepath.Join(engineRoot, "Assets", "wwwroot", "data", "images"),
                        filepath.Join(engineRoot, "Assets", "images"),
                }
                for _, td := range targetDirs {
                        _ = os.MkdirAll(td, 0755)
                        _ = os.MkdirAll(filepath.Join(td, "images"), 0755)
                        avPath := filepath.Join(td, "images", "av_tutorial.png")
                        if info, err := os.Stat(avPath); err != nil || info.Size() == 0 {
                                _ = os.WriteFile(avPath, placeholderPNG, 0644)
                        }
                        bgPath := filepath.Join(td, "images", "bg_tutorial.webp")
                        if _, err := os.Stat(bgPath); err != nil {
                                _ = os.WriteFile(bgPath, placeholderPNG, 0644)
                        }
                        bgPng := filepath.Join(td, "images", "bg_tutorial.png")
                        if _, err := os.Stat(bgPng); err != nil {
                                _ = os.WriteFile(bgPng, placeholderPNG, 0644)
                        }
                }
                for _, imgDir := range targetImageDirs {
                        _ = os.MkdirAll(imgDir, 0755)
                }
                // kopiuj JSONe
                filepath.Walk(srcData, func(p string, info fs.FileInfo, err error) error {
                        if err != nil || info.IsDir() {
                                return nil
                        }
                        if strings.HasSuffix(strings.ToLower(info.Name()), ".json") {
                                b, err := os.ReadFile(p)
                                if err == nil {
                                        for _, td := range targetDirs {
                                                _ = os.WriteFile(filepath.Join(td, info.Name()), b, 0644)
                                        }
                                        c++
                                }
                        }
                        return nil
                })
                // kopiuj OBRAZY z projektu do wszystkich wwwroot/images
                srcImages := filepath.Join(projectPath, "images")
                if _, err := os.Stat(srcImages); err == nil {
                        filepath.Walk(srcImages, func(p string, info fs.FileInfo, err error) error {
                                if err != nil || info.IsDir() {
                                        return nil
                                }
                                b, _ := os.ReadFile(p)
                                for _, imgDir := range targetImageDirs {
                                        _ = os.WriteFile(filepath.Join(imgDir, info.Name()), b, 0644)
                                }
                                // dodatkowo do wwwroot/images bezposrednio
                                _ = os.WriteFile(filepath.Join(engineRoot, "wwwroot", "images", info.Name()), b, 0644)
                                return nil
                        })
                } else {
                        // jesli projekt nie ma images, skopiuj z frontend/src/assets
                        fallbackAssets := filepath.Join(engineRoot, "..", "editor", "frontend", "src", "assets")
                        if _, err := os.Stat(fallbackAssets); err == nil {
                                filepath.Walk(fallbackAssets, func(p string, info fs.FileInfo, err error) error {
                                        if err != nil || info.IsDir() {
                                                return nil
                                        }
                                        low := strings.ToLower(info.Name())
                                        if strings.HasSuffix(low, ".webp") || strings.HasSuffix(low, ".png") || strings.HasSuffix(low, ".jpg") {
                                                b, _ := os.ReadFile(p)
                                                for _, imgDir := range targetImageDirs {
                                                        dst := filepath.Join(imgDir, info.Name())
                                                        if _, err := os.Stat(dst); os.IsNotExist(err) || func() bool { i, _ := os.Stat(dst); return i != nil && i.Size() < 100 }() {
                                                                _ = os.WriteFile(dst, b, 0644)
                                                        }
                                                }
                                        }
                                        return nil
                                })
                        }
                }
                if files, err := os.ReadDir(projectPath); err == nil {
                        for _, f := range files {
                                if f.IsDir() {
                                        continue
                                }
                                if strings.HasSuffix(strings.ToLower(f.Name()), ".janproj") {
                                        b, _ := os.ReadFile(filepath.Join(projectPath, f.Name()))
                                        for _, td := range targetDirs {
                                                _ = os.WriteFile(filepath.Join(td, "project.janproj"), b, 0644)
                                        }
                                        _ = os.WriteFile(filepath.Join(engineRoot, "wwwroot", "project.janproj"), b, 0644)
                                }
                        }
                }
                wwwData := filepath.Join(engineRoot, "wwwroot", "data")
                if _, err := os.Stat(filepath.Join(wwwData, "day1.json")); err != nil {
                        minimalDay := `[{"Id":"start","SceneTitle":"Manieczki - Start","Background":"","Text":"Start","Choices":[{"Text":"Dalej","Next":"koniec"}],"Type":"normal","Day":1},{"Id":"koniec","SceneTitle":"KONIEC","Background":"","Text":"Koniec","IsEndDay":true,"Type":"end_of_day","Day":1,"Choices":[]}]`
                        _ = os.WriteFile(filepath.Join(wwwData, "day1.json"), []byte(minimalDay), 0644)
                }
                if _, err := os.Stat(filepath.Join(wwwData, "project.janproj")); err != nil {
                        minimalProj := fmt.Sprintf(`{"gameName":"%s","author":"","version":"1.0.0","engineVersion":"2.0.0","startDay":"day1","startScene":"start"}`, filepath.Base(projectPath))
                        _ = os.WriteFile(filepath.Join(wwwData, "project.janproj"), []byte(minimalProj), 0644)
                        _ = os.WriteFile(filepath.Join(engineRoot, "wwwroot", "project.janproj"), []byte(minimalProj), 0644)
                }
                if _, err := os.Stat(filepath.Join(wwwData, "_manifest.json")); err != nil {
                        _ = os.WriteFile(filepath.Join(wwwData, "_manifest.json"), []byte(`{"days":["day1"],"startDay":"day1","startScene":"start","version":"2026-10-04T00:00:00.000Z"}`), 0644)
                }
                return c
        }

        // 4a. RESTORE
        runtime.LogInfo(a.ctx, "[PARKIET] dotnet restore...")
        restoreCmd := exec.Command("dotnet", "restore", enginePath, "--nologo")
        restoreCmd.Dir = engineRoot
        _ = restoreCmd.Run()

        // 4b. KOPIUJ PO RESTORE
        copied := copyData()
        runtime.LogInfo(a.ctx, fmt.Sprintf("[PARKIET] Skopiowano %d jsonow po restore", copied))
        wwwData := filepath.Join(engineRoot, "wwwroot", "data")
        if entries, _ := os.ReadDir(wwwData); entries != nil {
                for _, e := range entries {
                        runtime.LogInfo(a.ctx, "[PARKIET] wwwroot/data ma: "+e.Name())
                }
        }

        // 5. BUILD --no-restore
        runtime.LogInfo(a.ctx, "[PARKIET] dotnet build --no-restore start...")
        buildLogPath := filepath.Join(os.TempDir(), "kimonogm_build.log")
        buildLog, _ := os.Create(buildLogPath)
        buildArgs := []string{"build", enginePath, "-c", "Debug", "--no-restore", "--nologo"}
        buildCmd := exec.Command("dotnet", buildArgs...)
        buildCmd.Dir = engineRoot
        if buildLog != nil {
                buildCmd.Stdout = buildLog
                buildCmd.Stderr = buildLog
        }
        if err := buildCmd.Run(); err != nil {
                if buildLog != nil {
                        buildLog.Close()
                }
                b, _ := os.ReadFile(buildLogPath)
                tail := string(b)
                if len(tail) > 4000 {
                        tail = tail[len(tail)-4000:]
                }
                runtime.LogInfo(a.ctx, "[PARKIET] Build z SWA fail, probuje bez SWA...")
                buildLog2, _ := os.Create(buildLogPath)
                buildArgs2 := []string{"build", enginePath, "-c", "Debug", "--no-restore", "--nologo", "-p:StaticWebAssetsEnabled=false"}
                buildCmd2 := exec.Command("dotnet", buildArgs2...)
                buildCmd2.Dir = engineRoot
                if buildLog2 != nil {
                        buildCmd2.Stdout = buildLog2
                        buildCmd2.Stderr = buildLog2
                }
                if err2 := buildCmd2.Run(); err2 != nil {
                        if buildLog2 != nil {
                                buildLog2.Close()
                        }
                        b2, _ := os.ReadFile(buildLogPath)
                        tail2 := string(b2)
                        if len(tail2) > 4000 {
                                tail2 = tail2[len(tail2)-4000:]
                        }
                        return fmt.Errorf("dotnet build failed: %w\nLOG1:\n%s\n---TRY2---\n%s", err, tail, tail2)
                }
                if buildLog2 != nil {
                        buildLog2.Close()
                }
        } else {
                if buildLog != nil {
                        buildLog.Close()
                }
        }
        runtime.LogInfo(a.ctx, "[PARKIET] dotnet build OK")

        // 5b. KOPIUJ PONOWNIE PO BUILD
        copyData()

        // 6. KOPIUJ DO BIN
        var binWwwroots []string
        if matches, _ := filepath.Glob(filepath.Join(engineRoot, "bin", "*", "wwwroot")); matches != nil {
                binWwwroots = append(binWwwroots, matches...)
        }
        if matches, _ := filepath.Glob(filepath.Join(engineRoot, "bin", "*", "*", "wwwroot")); matches != nil {
                binWwwroots = append(binWwwroots, matches...)
        }
        binWwwroots = append(binWwwroots, filepath.Join(engineRoot, "bin", "Debug", "net8.0", "wwwroot"))
        seen := map[string]bool{}
        for _, wr := range binWwwroots {
                abs, _ := filepath.Abs(wr)
                if seen[abs] {
                        continue
                }
                seen[abs] = true
                if _, err := os.Stat(abs); err != nil {
                        continue
                }
                dataDir := filepath.Join(abs, "data")
                _ = os.MkdirAll(filepath.Join(dataDir, "images"), 0755)
                _ = os.MkdirAll(filepath.Join(abs, "images"), 0755)
                filepath.Walk(srcData, func(p string, info fs.FileInfo, err error) error {
                        if err != nil || info.IsDir() {
                                return nil
                        }
                        if strings.HasSuffix(strings.ToLower(info.Name()), ".json") {
                                b, _ := os.ReadFile(p)
                                _ = os.WriteFile(filepath.Join(dataDir, info.Name()), b, 0644)
                        }
                        return nil
                })
                if files, err := os.ReadDir(projectPath); err == nil {
                        for _, f := range files {
                                if f.IsDir() {
                                        continue
                                }
                                if strings.HasSuffix(strings.ToLower(f.Name()), ".janproj") {
                                        b, _ := os.ReadFile(filepath.Join(projectPath, f.Name()))
                                        _ = os.WriteFile(filepath.Join(abs, "project.janproj"), b, 0644)
                                        _ = os.WriteFile(filepath.Join(dataDir, "project.janproj"), b, 0644)
                                }
                        }
                }
                for _, av := range []string{
                        filepath.Join(abs, "images", "av_tutorial.png"),
                        filepath.Join(dataDir, "images", "av_tutorial.png"),
                } {
                        if info, err := os.Stat(av); err != nil || info.Size() == 0 {
                                _ = os.WriteFile(av, placeholderPNG, 0644)
                        }
                }
        }

        // 7. RUN --no-build
        exec.Command("cmd", "/C", "for /f \"tokens=5\" %a in ('netstat -aon ^| findstr :5020') do taskkill /F /PID %a").Run()
        time.Sleep(500 * time.Millisecond)
        for i := 0; i < 10; i++ {
                out, _ := exec.Command("cmd", "/C", "netstat -aon | findstr :5020").CombinedOutput()
                if len(out) == 0 {
                        break
                }
                time.Sleep(500 * time.Millisecond)
        }
        runtime.LogInfo(a.ctx, "[PARKIET] dotnet run --no-build :5020")
        gp := filepath.ToSlash(projectPath)
        prop := fmt.Sprintf("-p:GameAssetsRoot=\"%s\"", strings.ReplaceAll(gp, "\"", ""))
        runCmd := exec.Command("dotnet", "run", "--no-build", "--project", enginePath, prop, "--urls", "http://localhost:5020", "--no-launch-profile")
        runCmd.Dir = engineRoot
        runLogPath := filepath.Join(os.TempDir(), "kimonogm_run.log")
        runLog, _ := os.Create(runLogPath)
        if runLog != nil {
                runCmd.Stdout = runLog
                runCmd.Stderr = runLog
        }
        if err := runCmd.Start(); err != nil {
                return fmt.Errorf("dotnet run failed: %w", err)
        }
        time.Sleep(2000 * time.Millisecond)
        runtime.LogInfo(a.ctx, "[PARKIET] Uruchomiono na http://localhost:5020")
        return nil
}


func (a *App) BuildGame(projectPath string) (string, error) {
        if projectPath == "" {
                projectPath = a.projectPath
        }
        if projectPath == "" {
                return "", fmt.Errorf("brak sciezki projektu")
        }
        engineRoot, distRoot := a.findEngineRoot()
        if engineRoot == "" {
                return "", fmt.Errorf("nie znaleziono engine")
        }
        cmd := exec.Command("dotnet", "publish", "-c", "Release", "-o", distRoot, "-p:StaticWebAssetsCompressContent=false", "-p:StaticWebAssetCompressionMode=None", "-p:PublishTrimmed=false", "-p:BlazorWebAssemblyEnableTrimming=false")
        cmd.Dir = engineRoot
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
                return "", err
        }
        return distRoot, nil
}

// ---------- BUILD EXE (v56c863d tag logic) ----------

func (a *App) BuildExeGame(projectPath string) (string, error) {
        if projectPath == "" {
                projectPath = a.projectPath
        }
        if projectPath == "" {
                return "", fmt.Errorf("brak sciezki projektu")
        }
        a.projectPath = filepath.Clean(projectPath)

        engineRoot, distRoot := a.findEngineRoot()
        if engineRoot == "" {
                return "", fmt.Errorf("nie znaleziono engine")
        }

        // 0. placeholdery przed publish (tak samo jak w web-build)
        _ = os.MkdirAll(filepath.Join(engineRoot, "wwwroot", "data", "images"), 0755)
        _ = os.MkdirAll(filepath.Join(engineRoot, "wwwroot", "images"), 0755)
        _ = os.MkdirAll(filepath.Join(engineRoot, "Assets", "data", "images"), 0755)
        for _, p := range []string{
                filepath.Join(engineRoot, "wwwroot", "images", "av_tutorial.png"),
                filepath.Join(engineRoot, "wwwroot", "data", "images", "av_tutorial.png"),
                filepath.Join(engineRoot, "Assets", "data", "images", "av_tutorial.png"),
                filepath.Join(engineRoot, "wwwroot", "images", "bg_tutorial.webp"),
                filepath.Join(engineRoot, "Assets", "data", "images", "bg_tutorial.webp"),
        } {
                _ = os.MkdirAll(filepath.Dir(p), 0755)
                if info, err := os.Stat(p); err != nil || info.Size() == 0 {
                        _ = os.WriteFile(p, placeholderPNG, 0644)
                }
        }
        srcDataTmp := filepath.Join(projectPath, "Data")
        if _, err := os.Stat(srcDataTmp); os.IsNotExist(err) {
                srcDataTmp = projectPath
        }
        for _, td := range []string{filepath.Join(engineRoot, "Assets", "data"), filepath.Join(engineRoot, "Assets", "wwwroot", "data")} {
                _ = os.MkdirAll(td, 0755)
                filepath.Walk(srcDataTmp, func(p string, info fs.FileInfo, err error) error {
                        if err != nil || info.IsDir() {
                                return nil
                        }
                        if strings.HasSuffix(strings.ToLower(info.Name()), ".json") {
                                b, _ := os.ReadFile(p)
                                _ = os.WriteFile(filepath.Join(td, info.Name()), b, 0644)
                        }
                        return nil
                })
        }

        // kill old builds
        exec.Command("taskkill", "/F", "/IM", "KimonoGM.exe").Run()
        exec.Command("taskkill", "/F", "/IM", "MakerEngine.exe").Run()
        time.Sleep(500 * time.Millisecond)

        // build web dist first
        cmd := exec.Command("dotnet", "publish", "-c", "Release", "-o", distRoot, "-p:StaticWebAssetsCompressContent=false", "-p:StaticWebAssetCompressionMode=None", "-p:PublishTrimmed=false", "-p:BlazorWebAssemblyEnableTrimming=false")
        cmd.Dir = engineRoot
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
                return "", fmt.Errorf("publish failed: %w", err)
        }

        // copy game data to dist wwwroot
        srcData := filepath.Join(a.projectPath, "Data")
        if _, err := os.Stat(srcData); os.IsNotExist(err) {
                srcData = a.projectPath
        }
        dstData := filepath.Join(distRoot, "wwwroot", "data")
        os.MkdirAll(dstData, 0755)
        filepath.Walk(srcData, func(p string, info fs.FileInfo, err error) error {
                if err != nil || info.IsDir() {
                        return nil
                }
                if strings.HasSuffix(strings.ToLower(info.Name()), ".json") {
                        b, _ := os.ReadFile(p)
                        os.WriteFile(filepath.Join(dstData, info.Name()), b, 0644)
                }
                return nil
        })
        if _, err := os.Stat(filepath.Join(a.projectPath, "project.janproj")); err == nil {
                b, _ := os.ReadFile(filepath.Join(a.projectPath, "project.janproj"))
                os.WriteFile(filepath.Join(distRoot, "wwwroot", "data", "project.janproj"), b, 0644)
                os.WriteFile(filepath.Join(distRoot, "wwwroot", "project.janproj"), b, 0644)
        }

        srcImages := filepath.Join(a.projectPath, "images")
        dstImages := filepath.Join(distRoot, "wwwroot", "images")
        if _, err := os.Stat(srcImages); err == nil {
                os.MkdirAll(dstImages, 0755)
                filepath.Walk(srcImages, func(p string, info fs.FileInfo, err error) error {
                        if err != nil || info.IsDir() {
                                return nil
                        }
                        low := strings.ToLower(info.Name())
                        if strings.HasSuffix(low, ".webp") || strings.HasSuffix(low, ".png") || strings.HasSuffix(low, ".jpg") || strings.HasSuffix(low, ".jpeg") || strings.HasSuffix(low, ".gif") {
                                b, _ := os.ReadFile(p)
                                os.WriteFile(filepath.Join(dstImages, info.Name()), b, 0644)
                                os.WriteFile(filepath.Join(dstData, "images", info.Name()), b, 0644)
                        }
                        return nil
                })
        }

        // build EXE - engine Launcher (if exists) + auto-create if not
        downloads := getDownloadsFolder()
        gameName := filepath.Base(a.projectPath)
        safeName := strings.ReplaceAll(gameName, " ", "-")
        timestamp := time.Now().Format("20060102-1504")
        finalExe := filepath.Join(downloads, fmt.Sprintf("%s-EXE-%s.exe", safeName, timestamp))

        // szukaj Launchera w kilku miejscach
        launcherCandidates := []string{
                filepath.Join(filepath.Dir(engineRoot), "Launcher", "Launcher.csproj"),
                filepath.Join(engineRoot, "Launcher", "Launcher.csproj"),
                filepath.Join(filepath.Dir(engineRoot), "KimonoLauncher", "Launcher.csproj"),
        }
        var launcherProj string
        for _, c := range launcherCandidates {
                if _, err := os.Stat(c); err == nil {
                        launcherProj = c
                        break
                }
        }

        // jesli nie ma Launchera, stworz go na szybko w engine/Launcher - ZAWSZE nadpisz Program.cs z http serverem (fix file://)
        var launcherDir string
        if launcherProj == "" {
                launcherDir = filepath.Join(filepath.Dir(engineRoot), "Launcher")
                _ = os.MkdirAll(launcherDir, 0755)
                launcherProj = filepath.Join(launcherDir, "Launcher.csproj")
        } else {
                launcherDir = filepath.Dir(launcherProj)
        }
        // ZAWSZE nadpisz Program.cs i csproj zeby naprawic file:// -> http://localhost
        _ = os.MkdirAll(launcherDir, 0755)
        _ = os.Remove(filepath.Join(launcherDir, "Program.cs"))
        if true {
                        csprojContent := `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <OutputType>WinExe</OutputType>
    <TargetFramework>net8.0-windows</TargetFramework>
    <UseWindowsForms>true</UseWindowsForms>
    <ImplicitUsings>enable</ImplicitUsings>
    <PublishSingleFile>true</PublishSingleFile>
    <SelfContained>true</SelfContained>
    <RuntimeIdentifier>win-x64</RuntimeIdentifier>
    <IncludeNativeLibrariesForSelfExtract>true</IncludeNativeLibrariesForSelfExtract>
  </PropertyGroup>
  <ItemGroup>
    <PackageReference Include="Microsoft.Web.WebView2" Version="1.0.2739.15" />
  </ItemGroup>
</Project>`
                        _ = os.WriteFile(launcherProj, []byte(csprojContent), 0644)
                        programCs := `using System;
using System.IO;
using System.Linq;
using System.Net;
using System.Net.Sockets;
using System.Threading;
using System.Windows.Forms;
using Microsoft.Web.WebView2.WinForms;

class Program {
    static string GetMime(string path) {
        var ext = Path.GetExtension(path).ToLowerInvariant();
        return ext switch {
            ".html" => "text/html",
            ".js" => "application/javascript",
            ".css" => "text/css",
            ".json" => "application/json",
            ".png" => "image/png",
            ".jpg" or ".jpeg" => "image/jpeg",
            ".webp" => "image/webp",
            ".gif" => "image/gif",
            ".svg" => "image/svg+xml",
            ".wasm" => "application/wasm",
            ".woff" => "font/woff",
            ".woff2" => "font/woff2",
            ".ttf" => "font/ttf",
            ".ico" => "image/x-icon",
            _ => "application/octet-stream",
        };
    }

    [STAThread]
    static void Main() {
        Application.EnableVisualStyles();
        Application.SetCompatibleTextRenderingDefault(false);
        var baseDir = AppContext.BaseDirectory;
        var wwwroot = Path.Combine(baseDir, "wwwroot");
        if (!Directory.Exists(wwwroot) || !File.Exists(Path.Combine(wwwroot, "index.html"))) {
            var candidates = new[] {
                Path.Combine(baseDir, "wwwroot"),
                Path.Combine(baseDir, "..", "dist", "wwwroot"),
                Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "wwwroot"),
                Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), "Downloads", "wwwroot"),
            };
            foreach (var c in candidates) if (Directory.Exists(c) && File.Exists(Path.Combine(c, "index.html"))) { wwwroot = c; break; }
            if (!File.Exists(Path.Combine(wwwroot, "index.html"))) {
                try {
                    var dl = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), "Downloads");
                    if (Directory.Exists(dl)) {
                        foreach (var dir in Directory.GetDirectories(dl, "*-EXE-*").OrderByDescending(d => d)) {
                            var p = Path.Combine(dir, "wwwroot");
                            if (File.Exists(Path.Combine(p, "index.html"))) { wwwroot = p; break; }
                        }
                    }
                } catch {}
            }
            if (!File.Exists(Path.Combine(wwwroot, "index.html"))) {
                try {
                    foreach (var dir in Directory.GetDirectories(baseDir)) {
                        var p = Path.Combine(dir, "wwwroot");
                        if (File.Exists(Path.Combine(p, "index.html"))) { wwwroot = p; break; }
                        var p2 = Path.Combine(dir, "wwwroot", "wwwroot");
                        if (File.Exists(Path.Combine(p2, "index.html"))) { wwwroot = p2; break; }
                    }
                } catch {}
            }
        }

        if (!File.Exists(Path.Combine(wwwroot, "index.html"))) {
            MessageBox.Show("Brak wwwroot/index.html w " + wwwroot + "\nSzukano w:\n" + baseDir + "/wwwroot\nDownloads/wwwroot\nDownloads/*-EXE-*/wwwroot");
            return;
        }

        // znajdz wolny port
        int port = 0;
        var tcp = new TcpListener(IPAddress.Loopback, 0);
        tcp.Start();
        port = ((IPEndPoint)tcp.LocalEndpoint).Port;
        tcp.Stop();

        var http = new HttpListener();
        http.Prefixes.Add($"http://localhost:{port}/");
        http.Start();

        ThreadPool.QueueUserWorkItem(_ => {
            while (http.IsListening) {
                try {
                    var ctx = http.GetContext();
                    var req = ctx.Request;
                    var resp = ctx.Response;
                    var urlPath = req.Url.AbsolutePath.TrimStart('/');
                    if (string.IsNullOrEmpty(urlPath)) urlPath = "index.html";
                    // zabezpiecz przed .. 
                    var filePath = Path.GetFullPath(Path.Combine(wwwroot, urlPath.Replace('/', Path.DirectorySeparatorChar)));
                    if (!filePath.StartsWith(Path.GetFullPath(wwwroot), StringComparison.OrdinalIgnoreCase)) {
                        filePath = Path.Combine(wwwroot, "index.html");
                    }
                    if (!File.Exists(filePath)) {
                        // SPA fallback - jesli plik nie istnieje a nie ma rozszerzenia, zwroc index.html
                        if (!Path.HasExtension(filePath) || !File.Exists(filePath)) {
                            // sprobuj czy to nie jest _framework, css itp - jesli nie ma, to index.html
                            if (!File.Exists(filePath) && !urlPath.Contains(".")) {
                                filePath = Path.Combine(wwwroot, "index.html");
                            }
                        }
                    }
                    if (File.Exists(filePath)) {
                        var bytes = File.ReadAllBytes(filePath);
                        resp.ContentType = GetMime(filePath);
                        resp.ContentLength64 = bytes.Length;
                        resp.OutputStream.Write(bytes, 0, bytes.Length);
                    } else {
                        resp.StatusCode = 404;
                        var notFound = System.Text.Encoding.UTF8.GetBytes("404 Not Found: " + urlPath);
                        resp.ContentLength64 = notFound.Length;
                        resp.OutputStream.Write(notFound, 0, notFound.Length);
                    }
                    resp.OutputStream.Close();
                } catch { }
            }
        });

        var form = new Form() { Width=1280, Height=800, Text="Janusz Simulator 97 - Manieczki", StartPosition=FormStartPosition.CenterScreen };
        var webView = new WebView2() { Dock=DockStyle.Fill };
        form.Controls.Add(webView);
        form.FormClosed += (s,e) => { try { http.Stop(); } catch {} Environment.Exit(0); };
        form.Load += async (s,e) => {
            await webView.EnsureCoreWebView2Async();
            webView.CoreWebView2.Navigate($"http://localhost:{port}/");
        };
        Application.Run(form);
    }
}`;
                        _ = os.WriteFile(filepath.Join(launcherDir, "Program.cs"), []byte(programCs), 0644)
                }

        // publish Launcher
        tmpOut := filepath.Join(os.TempDir(), "kimono_launcher_publish")
        os.RemoveAll(tmpOut)
        c := exec.Command("dotnet", "publish", launcherProj, "-c", "Release", "-r", "win-x64", "--self-contained", "true", "-p:PublishSingleFile=true", "-p:IncludeNativeLibrariesForSelfExtract=true", "-o", tmpOut)
        c.Stdout = os.Stdout
        c.Stderr = os.Stderr
        if err := c.Run(); err == nil {
                // skopiuj wwwroot do folderu z exe
                srcWwwroot := filepath.Join(distRoot, "wwwroot")
                dstWwwroot := filepath.Join(tmpOut, "wwwroot")
                _ = os.MkdirAll(dstWwwroot, 0755)
                filepath.Walk(srcWwwroot, func(p string, info fs.FileInfo, err error) error {
                        if err != nil || info.IsDir() {
                                return nil
                        }
                        rel, _ := filepath.Rel(srcWwwroot, p)
                        dst := filepath.Join(dstWwwroot, rel)
                        _ = os.MkdirAll(filepath.Dir(dst), 0755)
                        b, _ := os.ReadFile(p)
                        _ = os.WriteFile(dst, b, 0644)
                        return nil
                })
                // znajdz Launcher.exe
                var exePath string
                if _, err := os.Stat(filepath.Join(tmpOut, "Launcher.exe")); err == nil {
                        exePath = filepath.Join(tmpOut, "Launcher.exe")
                } else {
                        // szukaj dowolnego exe
                        if files, _ := os.ReadDir(tmpOut); files != nil {
                                for _, f := range files {
                                        if strings.HasSuffix(strings.ToLower(f.Name()), ".exe") {
                                                exePath = filepath.Join(tmpOut, f.Name())
                                                break
                                        }
                                }
                        }
                }
                if exePath != "" {
                        // finalny folder z gra + exe
                        finalFolder := filepath.Join(downloads, fmt.Sprintf("%s-EXE-%s", safeName, timestamp))
                        _ = os.MkdirAll(finalFolder, 0755)
                        // skopiuj exe jako Manieczki.exe
                        b, _ := os.ReadFile(exePath)
                        _ = os.WriteFile(filepath.Join(finalFolder, fmt.Sprintf("%s.exe", safeName)), b, 0644)
                        // skopiuj wwwroot
                        filepath.Walk(tmpOut, func(p string, info fs.FileInfo, err error) error {
                                if err != nil || info.IsDir() {
                                        return nil
                                }
                                rel, _ := filepath.Rel(tmpOut, p)
                                if rel == filepath.Base(exePath) {
                                        return nil
                                }
                                dst := filepath.Join(finalFolder, rel)
                                _ = os.MkdirAll(filepath.Dir(dst), 0755)
                                bb, _ := os.ReadFile(p)
                                _ = os.WriteFile(dst, bb, 0644)
                                return nil
                        })
                        // zip calego folderu
                        finalZip := filepath.Join(downloads, fmt.Sprintf("%s-EXE-%s.zip", safeName, timestamp))
                        zf, _ := os.Create(finalZip)
                        zw := zip.NewWriter(zf)
                        filepath.Walk(finalFolder, func(p string, info fs.FileInfo, err error) error {
                                if err != nil || info.IsDir() {
                                        return nil
                                }
                                rel, _ := filepath.Rel(finalFolder, p)
                                f, _ := zw.Create(filepath.ToSlash(rel))
                                src, _ := os.Open(p)
                                if src != nil {
                                        io.Copy(f, src)
                                        src.Close()
                                }
                                return nil
                        })
                        zw.Close()
                        zf.Close()
                        // skopiuj exe osobno do Downloads + wwwroot obok niego zeby dzialal z Downloads root
                        _ = os.WriteFile(finalExe, b, 0644)
                        // kopia wwwroot do Downloads/wwwroot dla exe w root
                        dlWwwroot := filepath.Join(downloads, "wwwroot")
                        _ = os.MkdirAll(dlWwwroot, 0755)
                        srcWwwroot := filepath.Join(distRoot, "wwwroot")
                        filepath.Walk(srcWwwroot, func(p string, info fs.FileInfo, err error) error {
                                if err != nil || info.IsDir() {
                                        return nil
                                }
                                rel, _ := filepath.Rel(srcWwwroot, p)
                                dst := filepath.Join(dlWwwroot, rel)
                                _ = os.MkdirAll(filepath.Dir(dst), 0755)
                                bb, _ := os.ReadFile(p)
                                _ = os.WriteFile(dst, bb, 0644)
                                return nil
                        })
                        exec.Command("explorer", "/select,", finalExe).Start()
                        return finalExe, nil
                }
        }

        // fallback - zip dist jako exe placeholder jesli Launcher nie zbudowany
        finalExe = filepath.Join(downloads, fmt.Sprintf("%s-EXE-%s.zip", safeName, timestamp))
        zf, _ := os.Create(finalExe)
        zw := zip.NewWriter(zf)
        wwwrootPath := filepath.Join(distRoot, "wwwroot")
        filepath.Walk(wwwrootPath, func(p string, info fs.FileInfo, err error) error {
                if err != nil || info.IsDir() {
                        return nil
                }
                rel, _ := filepath.Rel(wwwrootPath, p)
                f, _ := zw.Create(filepath.ToSlash(rel))
                src, _ := os.Open(p)
                if src != nil {
                        io.Copy(f, src)
                        src.Close()
                }
                return nil
        })
        zw.Close()
        zf.Close()

        exec.Command("explorer", "/select,", finalExe).Start()
        return finalExe, nil
}

// ---------- BUILD PUBLIC PACK -> DOWNLOADS ----------

func (a *App) BuildPublicPack(projectPath string) (string, error) {
        if projectPath == "" {
                projectPath = a.projectPath
        }
        if projectPath == "" {
                return "", fmt.Errorf("brak sciezki projektu")
        }
        projectPath = filepath.Clean(projectPath)
        a.projectPath = projectPath

        engineRoot, distRoot := a.findEngineRoot()
        if engineRoot == "" {
                return "", fmt.Errorf("nie znaleziono engine/MakerEngine.csproj")
        }
        os.MkdirAll(distRoot, 0755)

        // 0. FIX placeholdery przed publish - inaczej StaticWebAssets wywala "No file exists"
        _ = os.MkdirAll(filepath.Join(engineRoot, "wwwroot", "data", "images"), 0755)
        _ = os.MkdirAll(filepath.Join(engineRoot, "wwwroot", "images"), 0755)
        _ = os.MkdirAll(filepath.Join(engineRoot, "Assets", "data", "images"), 0755)
        _ = os.MkdirAll(filepath.Join(engineRoot, "Assets", "wwwroot", "data", "images"), 0755)
        for _, p := range []string{
                filepath.Join(engineRoot, "wwwroot", "images", "av_tutorial.png"),
                filepath.Join(engineRoot, "wwwroot", "data", "images", "av_tutorial.png"),
                filepath.Join(engineRoot, "Assets", "data", "images", "av_tutorial.png"),
                filepath.Join(engineRoot, "Assets", "wwwroot", "data", "images", "av_tutorial.png"),
                filepath.Join(engineRoot, "wwwroot", "images", "bg_tutorial.webp"),
                filepath.Join(engineRoot, "wwwroot", "data", "images", "bg_tutorial.webp"),
                filepath.Join(engineRoot, "Assets", "data", "images", "bg_tutorial.webp"),
        } {
                _ = os.MkdirAll(filepath.Dir(p), 0755)
                if info, err := os.Stat(p); err != nil || info.Size() == 0 {
                        _ = os.WriteFile(p, placeholderPNG, 0644)
                }
        }
        // skopiuj Data do Assets/data zeby COPIED FROM Assets -> wwwroot mialo co kopiowac
        srcDataTmp := filepath.Join(projectPath, "Data")
        if _, err := os.Stat(srcDataTmp); os.IsNotExist(err) {
                srcDataTmp = projectPath
        }
        for _, td := range []string{filepath.Join(engineRoot, "Assets", "data"), filepath.Join(engineRoot, "Assets", "wwwroot", "data")} {
                _ = os.MkdirAll(td, 0755)
                filepath.Walk(srcDataTmp, func(p string, info fs.FileInfo, err error) error {
                        if err != nil || info.IsDir() {
                                return nil
                        }
                        if strings.HasSuffix(strings.ToLower(info.Name()), ".json") {
                                b, _ := os.ReadFile(p)
                                _ = os.WriteFile(filepath.Join(td, info.Name()), b, 0644)
                        }
                        return nil
                })
                if _, err := os.Stat(filepath.Join(projectPath, "project.janproj")); err == nil {
                        b, _ := os.ReadFile(filepath.Join(projectPath, "project.janproj"))
                        _ = os.WriteFile(filepath.Join(td, "project.janproj"), b, 0644)
                }
                if _, err := os.Stat(filepath.Join(td, "day1.json")); err != nil {
                        _ = os.WriteFile(filepath.Join(td, "day1.json"), []byte(`[{"Id":"start","SceneTitle":"Start","Background":"","Text":"Start","Choices":[{"Text":"Dalej","Next":"koniec"}],"Type":"normal","Day":1},{"Id":"koniec","SceneTitle":"KONIEC","Background":"","Text":"Koniec","IsEndDay":true,"Type":"end_of_day","Day":1,"Choices":[]}]`), 0644)
                }
                if _, err := os.Stat(filepath.Join(td, "_manifest.json")); err != nil {
                        _ = os.WriteFile(filepath.Join(td, "_manifest.json"), []byte(`{"days":["day1"],"startDay":"day1","startScene":"start"}`), 0644)
                }
        }

        // 1. dotnet publish
        cmd := exec.Command("dotnet", "publish", "-c", "Release", "-o", distRoot, "-p:StaticWebAssetsCompressContent=false", "-p:StaticWebAssetCompressionMode=None", "-p:PublishTrimmed=false", "-p:BlazorWebAssemblyEnableTrimming=false")
        cmd.Dir = engineRoot
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
                return "", fmt.Errorf("dotnet publish failed: %w", err)
        }

        // 2. copy Data + images + sounds + project.janproj
        srcData := filepath.Join(projectPath, "Data")
        if _, err := os.Stat(srcData); os.IsNotExist(err) {
                srcData = projectPath
        }
        dstData := filepath.Join(distRoot, "wwwroot", "data")
        os.MkdirAll(dstData, 0755)
        if files, err := os.ReadDir(dstData); err == nil {
                for _, f := range files {
                        if strings.HasSuffix(strings.ToLower(f.Name()), ".json") {
                                os.Remove(filepath.Join(dstData, f.Name()))
                        }
                }
        }
        filepath.Walk(srcData, func(p string, info fs.FileInfo, err error) error {
                if err != nil || info.IsDir() {
                        return nil
                }
                if strings.HasSuffix(strings.ToLower(info.Name()), ".json") {
                        b, _ := os.ReadFile(p)
                        os.WriteFile(filepath.Join(dstData, info.Name()), b, 0644)
                }
                return nil
        })

        srcImages := filepath.Join(projectPath, "images")
        dstImages := filepath.Join(distRoot, "wwwroot", "images")
        if _, err := os.Stat(srcImages); err == nil {
                os.MkdirAll(dstImages, 0755)
                filepath.Walk(srcImages, func(p string, info fs.FileInfo, err error) error {
                        if err != nil || info.IsDir() {
                                return nil
                        }
                        low := strings.ToLower(info.Name())
                        if strings.HasSuffix(low, ".webp") || strings.HasSuffix(low, ".png") || strings.HasSuffix(low, ".jpg") || strings.HasSuffix(low, ".jpeg") || strings.HasSuffix(low, ".gif") {
                                b, _ := os.ReadFile(p)
                                os.WriteFile(filepath.Join(dstImages, info.Name()), b, 0644)
                        }
                        return nil
                })
        }

        srcSounds := filepath.Join(projectPath, "sounds")
        dstSounds := filepath.Join(distRoot, "wwwroot", "sounds")
        if _, err := os.Stat(srcSounds); err == nil {
                filepath.Walk(srcSounds, func(p string, info fs.FileInfo, err error) error {
                        if err != nil || info.IsDir() {
                                return nil
                        }
                        rel, _ := filepath.Rel(srcSounds, p)
                        dst := filepath.Join(dstSounds, rel)
                        os.MkdirAll(filepath.Dir(dst), 0755)
                        b, _ := os.ReadFile(p)
                        os.WriteFile(dst, b, 0644)
                        return nil
                })
        }

        if _, err := os.Stat(filepath.Join(projectPath, "project.janproj")); err == nil {
                b, _ := os.ReadFile(filepath.Join(projectPath, "project.janproj"))
                os.WriteFile(filepath.Join(distRoot, "wwwroot", "project.janproj"), b, 0644)
        }

        // 3. ZIP + unpacked folder to Downloads
        gameName := filepath.Base(projectPath)
        safeName := strings.ReplaceAll(gameName, " ", "-")
        timestamp := time.Now().Format("20060102-1504")

        downloads := getDownloadsFolder()
        os.MkdirAll(downloads, 0755)

        finalZip := filepath.Join(downloads, fmt.Sprintf("%s-PUBLIC-%s.zip", safeName, timestamp))
        unpackedFolder := filepath.Join(downloads, fmt.Sprintf("%s-PUBLIC-%s", safeName, timestamp))
        os.RemoveAll(unpackedFolder)

        zf, err := os.Create(finalZip)
        if err != nil {
                return "", err
        }
        zw := zip.NewWriter(zf)

        wwwrootPath := filepath.Join(distRoot, "wwwroot")
        filepath.Walk(wwwrootPath, func(p string, info fs.FileInfo, err error) error {
                if err != nil || info.IsDir() {
                        return nil
                }
                rel, _ := filepath.Rel(wwwrootPath, p)
                relSlash := filepath.ToSlash(rel)

                // do ZIP
                f, _ := zw.Create(relSlash)
                src, _ := os.Open(p)
                if src != nil {
                        io.Copy(f, src)
                        src.Close()
                }

                // do folderu rozpakowanego
                dstFile := filepath.Join(unpackedFolder, rel)
                os.MkdirAll(filepath.Dir(dstFile), 0755)
                b, _ := os.ReadFile(p)
                os.WriteFile(dstFile, b, 0644)

                return nil
        })

        zw.Close()
        zf.Close()

        exec.Command("explorer", "/select,", finalZip).Start()
        return finalZip, nil
}

// alias for old button
func (a *App) ExportPack(projectPath string) (string, error) {
        return a.BuildPublicPack(projectPath)
}

// ---------- HOST ----------
// Teraz HOST = BUILD PUBLIC PACK do Downloads (BEZ LOCALHOSTA)
func (a *App) HostGame(projectPath string, port int) (string, error) {
        return a.BuildPublicPack(projectPath)
}
func (a *App) StopHost() error {
        a.hostMu.Lock()
        defer a.hostMu.Unlock()
        if a.hostServer != nil {
                err := a.hostServer.Close()
                a.hostServer = nil
                a.hostPort = 0
                return err
        }
        return nil
}
func (a *App) GetHostStatus() (bool, int, string) {
        a.hostMu.Lock()
        defer a.hostMu.Unlock()
        if a.hostServer != nil && a.hostPort != 0 {
                return true, a.hostPort, fmt.Sprintf("http://localhost:%d", a.hostPort)
        }
        return false, 0, ""
}
