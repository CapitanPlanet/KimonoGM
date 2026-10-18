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

        // Znajdź builder - tools/BuilderEXE
        builderDir := `C:\Users\MD-Core\KimonoGM\tools\BuilderEXE`
        if _, err := os.Stat(builderDir); err != nil {
                // fallback szukanie
                candidates := []string{
                        filepath.Join("..", "tools", "BuilderEXE"),
                        filepath.Join("..", "..", "tools", "BuilderEXE"),
                        filepath.Join("tools", "BuilderEXE"),
                }
                for _, c := range candidates {
                        abs, _ := filepath.Abs(c)
                        if _, err := os.Stat(abs); err == nil {
                                builderDir = abs
                                break
                        }
                }
        }

        fmt.Printf("[EDITOR] Budowanie %s -> %s via %s\n", projectPath, outExe, builderDir)

        // go run Builder-Main.go --game X --out Y
        cmd := exec.Command("go", "run", "Builder-Main.go", "--game", projectPath, "--out", outExe)
        cmd.Dir = builderDir
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr

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

// === TU JEST NAPRAWA - BuildExeGame v37 - REAL EXE BUILDER ===
func (a *App) BuildExeGame(projectPath string) (string, error) {
        logPath := `C:\Users\MD-Core\KimonoGM\build-exe.log`
        _ = os.WriteFile(logPath, []byte(fmt.Sprintf("START: %s\nTIME: %s\n", projectPath, time.Now().Format(time.RFC3339))), 0644)

        // BEZWZGLĘDNA ŚCIEŻKA DO BUILDERA - nie licz na Getwd
        builderDir := `C:\Users\MD-Core\KimonoGM\tools\BuilderEXE`
        builderMain := filepath.Join(builderDir, "Builder-Main.go")

        // TASKKILL v2..v50 + build-server shutdown (v36 fix)
        if goruntime.GOOS == "windows" {
                exec.Command("taskkill", "/F", "/IM", "qqppppwww.exe").Run()
                for i := 2; i <= 50; i++ {
                        exec.Command("taskkill", "/F", "/IM", fmt.Sprintf("qqppppwww_v%d.exe", i)).Run()
                }
                exec.Command("dotnet", "build-server", "shutdown").Run()
                time.Sleep(500 * time.Millisecond)
        }

        cmd := exec.Command("go", "run", builderMain, "--game", projectPath)
        cmd.Dir = builderDir
        out, err := cmd.CombinedOutput()

        // zapisz log
        _ = os.WriteFile(logPath, append([]byte(fmt.Sprintf("CMD: go run %s --game %s\nDIR: %s\n\n", builderMain, projectPath, builderDir)), out...), 0644)
        if err != nil {
                _ = os.WriteFile(logPath, append(out, []byte(fmt.Sprintf("\nEXIT ERR: %v\n", err))...), 0644)
        }

        // log do wails
        if a.ctx != nil {
                runtime.LogInfo(a.ctx, fmt.Sprintf("[BuildExeGame] %s", string(out)))
                if err != nil {
                        runtime.LogError(a.ctx, fmt.Sprintf("[BuildExeGame] ERR: %v", err))
                }
        }
        fmt.Printf("[BuildExeGame] %s\n", string(out))

        if err != nil {
                return string(out), fmt.Errorf("build failed: %w\n%s", err, string(out))
        }
        return string(out), nil
}

// wrapper dla kompatybilności - stary BUILDER też dostępny
func (a *App) BuildExeGameLegacy(projectPath string) (string, error) {
        return a.BuildExeGame(projectPath)
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

// ---------- BUILD EXE + PUBLIC PACK - dodane do maina, nie ruszają CreateProject ----------
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

func (a *App) getDownloadsFolder() string {
        return getDownloadsFolder()
}
func (a *App) findEngineRoot() (string, string) {
        cwd, _ := os.Getwd()
        candidates := []string{
                filepath.Join(filepath.Dir(filepath.Dir(cwd)), "engine"),
                filepath.Join(cwd, "..", "..", "engine"),
                `C:\Users\MD-Core\KimonoGM\engine`,
                filepath.Join(cwd, "..", "engine"),
                "engine",
        }
        for _, c := range candidates {
                abs, _ := filepath.Abs(c)
                if _, err := os.Stat(filepath.Join(abs, "MakerEngine.csproj")); err == nil {
                        return abs, filepath.Join(filepath.Dir(abs), "dist")
                }
        }
        return "", ""
}

func (a *App) BuildExeWebPackage(projectPath string) (string, error) {
        if projectPath == "" {
                projectPath = a.projectPath
        }
        if projectPath == "" {
                return "", fmt.Errorf("brak sciezki projektu")
        }
        a.projectPath = filepath.Clean(projectPath)
        engineRoot, distRoot := a.findEngineRoot()
        if engineRoot == "" {
                return "", fmt.Errorf("nie znaleziono engine/MakerEngine.csproj")
        }
        exec.Command("taskkill", "/F", "/IM", "KimonoGM.exe").Run()
        exec.Command("taskkill", "/F", "/IM", "MakerEngine.exe").Run()
        time.Sleep(500 * time.Millisecond)

        cmd := exec.Command("dotnet", "publish", "-c", "Release", "-o", distRoot, "-p:StaticWebAssetsCompressContent=false", "-p:PublishTrimmed=false", "-p:BlazorWebAssemblyEnableTrimming=false")
        cmd.Dir = engineRoot
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
                return "", fmt.Errorf("dotnet publish failed: %w", err)
        }

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
        // FIX janproj
        if files, err := os.ReadDir(a.projectPath); err == nil {
                for _, f := range files {
                        if f.IsDir() {
                                continue
                        }
                        if strings.HasSuffix(strings.ToLower(f.Name()), ".janproj") {
                                b, _ := os.ReadFile(filepath.Join(a.projectPath, f.Name()))
                                os.WriteFile(filepath.Join(dstData, f.Name()), b, 0644)
                                os.WriteFile(filepath.Join(dstData, "project.janproj"), b, 0644)
                                os.WriteFile(filepath.Join(distRoot, "wwwroot", f.Name()), b, 0644)
                                os.WriteFile(filepath.Join(distRoot, "wwwroot", "project.janproj"), b, 0644)
                        }
                }
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
                        }
                        return nil
                })
        }

        downloads := getDownloadsFolder()
        safeName := strings.ReplaceAll(filepath.Base(a.projectPath), " ", "-")
        timestamp := time.Now().Format("20060102-1504")
        finalZip := filepath.Join(downloads, fmt.Sprintf("%s-EXE-%s.zip", safeName, timestamp))
        zf, err := os.Create(finalZip)
        if err != nil {
                return "", err
        }
        zw := zip.NewWriter(zf)
        wwwroot := filepath.Join(distRoot, "wwwroot")
        filepath.Walk(wwwroot, func(p string, info fs.FileInfo, err error) error {
                if err != nil || info.IsDir() {
                        return nil
                }
                rel, _ := filepath.Rel(wwwroot, p)
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
        exec.Command("explorer", "/select,", finalZip).Start()
        return finalZip, nil
}

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

        cmd := exec.Command("dotnet", "publish", "-c", "Release", "-o", distRoot, "-p:StaticWebAssetsCompressContent=false", "-p:PublishTrimmed=false", "-p:BlazorWebAssemblyEnableTrimming=false")
        cmd.Dir = engineRoot
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
                return "", fmt.Errorf("dotnet publish failed: %w", err)
        }

        srcData := filepath.Join(projectPath, "Data")
        if _, err := os.Stat(srcData); os.IsNotExist(err) {
                srcData = projectPath
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
        // kopiuj project.janproj i *.janproj - FIX avatarów i statów
        if files, err := os.ReadDir(projectPath); err == nil {
                for _, f := range files {
                        if f.IsDir() {
                                continue
                        }
                        ln := strings.ToLower(f.Name())
                        if strings.HasSuffix(ln, ".janproj") {
                                b, _ := os.ReadFile(filepath.Join(projectPath, f.Name()))
                                os.WriteFile(filepath.Join(dstData, f.Name()), b, 0644)
                                os.WriteFile(filepath.Join(distRoot, "wwwroot", f.Name()), b, 0644)
                                os.WriteFile(filepath.Join(distRoot, "wwwroot", "data", f.Name()), b, 0644)
                                if ln != "project.janproj" {
                                        os.WriteFile(filepath.Join(dstData, "project.janproj"), b, 0644)
                                        os.WriteFile(filepath.Join(distRoot, "wwwroot", "project.janproj"), b, 0644)
                                }
                        }
                }
        }

        for _, pair := range [][2]string{
                {filepath.Join(projectPath, "images"), filepath.Join(distRoot, "wwwroot", "images")},
                {filepath.Join(projectPath, "sounds"), filepath.Join(distRoot, "wwwroot", "sounds")},
        } {
                if _, err := os.Stat(pair[0]); err == nil {
                        filepath.Walk(pair[0], func(p string, info fs.FileInfo, err error) error {
                                if err != nil || info.IsDir() {
                                        return nil
                                }
                                rel, _ := filepath.Rel(pair[0], p)
                                dst := filepath.Join(pair[1], rel)
                                os.MkdirAll(filepath.Dir(dst), 0755)
                                b, _ := os.ReadFile(p)
                                os.WriteFile(dst, b, 0644)
                                fallback := filepath.Join(distRoot, "wwwroot", "data", "images", rel)
                                os.MkdirAll(filepath.Dir(fallback), 0755)
                                os.WriteFile(fallback, b, 0644)
                                return nil
                        })
                }
        }

        safeName := strings.ReplaceAll(filepath.Base(projectPath), " ", "-")
        timestamp := time.Now().Format("20060102-1504")
        downloads := getDownloadsFolder()
        os.MkdirAll(downloads, 0755)
        finalZip := filepath.Join(downloads, fmt.Sprintf("%s-PUBLIC-%s.zip", safeName, timestamp))
        unpacked := filepath.Join(downloads, fmt.Sprintf("%s-PUBLIC-%s", safeName, timestamp))
        os.RemoveAll(unpacked)

        zf, err := os.Create(finalZip)
        if err != nil {
                return "", err
        }
        zw := zip.NewWriter(zf)
        wwwroot := filepath.Join(distRoot, "wwwroot")
        filepath.Walk(wwwroot, func(p string, info fs.FileInfo, err error) error {
                if err != nil || info.IsDir() {
                        return nil
                }
                rel, _ := filepath.Rel(wwwroot, p)
                f, _ := zw.Create(filepath.ToSlash(rel))
                src, _ := os.Open(p)
                if src != nil {
                        io.Copy(f, src)
                        src.Close()
                }
                dstFile := filepath.Join(unpacked, rel)
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

func (a *App) ExportPack(projectPath string) (string, error) {
        return a.BuildPublicPack(projectPath)
}

// ========== HOST NA LOCALHOŚCIE - PROSTE ==========
func (a *App) StopHost() error {
        a.hostMu.Lock()
        defer a.hostMu.Unlock()
        if a.hostServer != nil {
                err := a.hostServer.Close()
                a.hostServer = nil
                a.hostPort = 0
                if a.ctx != nil {
                        runtime.LogInfo(a.ctx, "[HOST] Zatrzymano")
                }
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

func (a *App) HostGame(projectPath string, port int) (string, error) {
        if projectPath == "" {
                projectPath = a.projectPath
        }
        if projectPath == "" {
                return "", fmt.Errorf("brak sciezki projektu")
        }
        projectPath = filepath.Clean(projectPath)
        a.projectPath = projectPath

        if port == 0 {
                port = 8080
        }

        // zatrzymaj poprzedni host
        _ = a.StopHost()

        engineRoot, distRoot := a.findEngineRoot()
        if engineRoot == "" {
                return "", fmt.Errorf("nie znaleziono engine/MakerEngine.csproj")
        }
        os.MkdirAll(distRoot, 0755)

        if a.ctx != nil {
                runtime.LogInfo(a.ctx, fmt.Sprintf("[HOST] Budowanie %s", projectPath))
        }

        // to samo co BuildPublicPack - publish
        cmd := exec.Command("dotnet", "publish", "-c", "Release", "-o", distRoot, "-p:StaticWebAssetsCompressContent=false", "-p:PublishTrimmed=false", "-p:BlazorWebAssemblyEnableTrimming=false")
        cmd.Dir = engineRoot
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
                return "", fmt.Errorf("dotnet publish failed: %w", err)
        }

        // skopiuj dane projektu - dokładnie jak w paczce
        srcData := filepath.Join(projectPath, "Data")
        if _, err := os.Stat(srcData); os.IsNotExist(err) {
                srcData = projectPath
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
        // kopiuj project.janproj i *.janproj z roota projektu - TU BYŁ BUG z avatarami i statami
        // kopiujemy do BOTH wwwroot/data i wwwroot (root) bo engine raz bierze stąd raz stamtąd
        if files, err := os.ReadDir(projectPath); err == nil {
                for _, f := range files {
                        if f.IsDir() {
                                continue
                        }
                        ln := strings.ToLower(f.Name())
                        if strings.HasSuffix(ln, ".janproj") || ln == "project.janproj" {
                                b, _ := os.ReadFile(filepath.Join(projectPath, f.Name()))
                                os.WriteFile(filepath.Join(dstData, f.Name()), b, 0644)
                                os.WriteFile(filepath.Join(distRoot, "wwwroot", f.Name()), b, 0644)
                                os.WriteFile(filepath.Join(distRoot, "wwwroot", "data", f.Name()), b, 0644)
                                // też jako project.janproj jeśli inna nazwa
                                if ln != "project.janproj" {
                                        os.WriteFile(filepath.Join(dstData, "project.janproj"), b, 0644)
                                        os.WriteFile(filepath.Join(distRoot, "wwwroot", "project.janproj"), b, 0644)
                                }
                        }
                }
        }

        for _, pair := range [][2]string{
                {filepath.Join(projectPath, "images"), filepath.Join(distRoot, "wwwroot", "images")},
                {filepath.Join(projectPath, "sounds"), filepath.Join(distRoot, "wwwroot", "sounds")},
        } {
                if _, err := os.Stat(pair[0]); err == nil {
                        filepath.Walk(pair[0], func(p string, info fs.FileInfo, err error) error {
                                if err != nil || info.IsDir() {
                                        return nil
                                }
                                rel, _ := filepath.Rel(pair[0], p)
                                dst := filepath.Join(pair[1], rel)
                                os.MkdirAll(filepath.Dir(dst), 0755)
                                b, _ := os.ReadFile(p)
                                os.WriteFile(dst, b, 0644)
                                // kopia do data/images jako fallback
                                fallback := filepath.Join(distRoot, "wwwroot", "data", "images", rel)
                                os.MkdirAll(filepath.Dir(fallback), 0755)
                                os.WriteFile(fallback, b, 0644)
                                return nil
                        })
                }
        }

        wwwroot := filepath.Join(distRoot, "wwwroot")

        // ===== PRZEPIĘKNA PACZKA ZIP - jak wcześniej =====
        safeName := strings.ReplaceAll(filepath.Base(projectPath), " ", "-")
        timestamp := time.Now().Format("20060102-1504")
        downloads := getDownloadsFolder()
        os.MkdirAll(downloads, 0755)
        finalZip := filepath.Join(downloads, fmt.Sprintf("%s-HOST-%s.zip", safeName, timestamp))
        unpacked := filepath.Join(downloads, fmt.Sprintf("%s-HOST-%s", safeName, timestamp))
        os.RemoveAll(unpacked)

        zf, errZip := os.Create(finalZip)
        if errZip == nil {
                zw := zip.NewWriter(zf)
                filepath.Walk(wwwroot, func(p string, info fs.FileInfo, err error) error {
                        if err != nil || info.IsDir() {
                                return nil
                        }
                        rel, _ := filepath.Rel(wwwroot, p)
                        f, _ := zw.Create(filepath.ToSlash(rel))
                        src, _ := os.Open(p)
                        if src != nil {
                                io.Copy(f, src)
                                src.Close()
                        }
                        dstFile := filepath.Join(unpacked, rel)
                        os.MkdirAll(filepath.Dir(dstFile), 0755)
                        b, _ := os.ReadFile(p)
                        os.WriteFile(dstFile, b, 0644)
                        return nil
                })
                zw.Close()
                zf.Close()
                if a.ctx != nil {
                        runtime.LogInfo(a.ctx, fmt.Sprintf("[HOST] PACZKA: %s", finalZip))
                }
        }

        // === KONIEC - TYLKO PACZKA DO POBRANYCH, BEZ LOCALHOSTA - DEFINITYWNIE ===
        // localhost wywalony bo cache Blazora jebał avatary i staty
        if a.ctx != nil {
                runtime.LogInfo(a.ctx, fmt.Sprintf("[HOST] PACZKA GOTOWA (bez localhosta): %s", finalZip))
        }
        exec.Command("explorer", "/select,", finalZip).Start()
        return finalZip, nil
}

// NOWA FUNKCJA: tylko kopiuje do Pobranych, nie odpala serwera - dla przycisku bez cache problemu
func (a *App) CopyToDownloadsOnly(projectPath string) (string, error) {
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

        if a.ctx != nil {
                runtime.LogInfo(a.ctx, fmt.Sprintf("[COPY] Budowanie %s -> Pobrane", projectPath))
        }

        cmd := exec.Command("dotnet", "publish", "-c", "Release", "-o", distRoot, "-p:StaticWebAssetsCompressContent=false", "-p:PublishTrimmed=false", "-p:BlazorWebAssemblyEnableTrimming=false")
        cmd.Dir = engineRoot
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
                return "", fmt.Errorf("dotnet publish failed: %w", err)
        }

        srcData := filepath.Join(projectPath, "Data")
        if _, err := os.Stat(srcData); os.IsNotExist(err) {
                srcData = projectPath
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
        if files, err := os.ReadDir(projectPath); err == nil {
                for _, f := range files {
                        if f.IsDir() {
                                continue
                        }
                        ln := strings.ToLower(f.Name())
                        if strings.HasSuffix(ln, ".janproj") {
                                b, _ := os.ReadFile(filepath.Join(projectPath, f.Name()))
                                os.WriteFile(filepath.Join(dstData, f.Name()), b, 0644)
                                os.WriteFile(filepath.Join(distRoot, "wwwroot", f.Name()), b, 0644)
                                os.WriteFile(filepath.Join(distRoot, "wwwroot", "data", f.Name()), b, 0644)
                                if ln != "project.janproj" {
                                        os.WriteFile(filepath.Join(dstData, "project.janproj"), b, 0644)
                                        os.WriteFile(filepath.Join(distRoot, "wwwroot", "project.janproj"), b, 0644)
                                }
                        }
                }
        }

        for _, pair := range [][2]string{
                {filepath.Join(projectPath, "images"), filepath.Join(distRoot, "wwwroot", "images")},
                {filepath.Join(projectPath, "sounds"), filepath.Join(distRoot, "wwwroot", "sounds")},
        } {
                if _, err := os.Stat(pair[0]); err == nil {
                        filepath.Walk(pair[0], func(p string, info fs.FileInfo, err error) error {
                                if err != nil || info.IsDir() {
                                        return nil
                                }
                                rel, _ := filepath.Rel(pair[0], p)
                                dst := filepath.Join(pair[1], rel)
                                os.MkdirAll(filepath.Dir(dst), 0755)
                                b, _ := os.ReadFile(p)
                                os.WriteFile(dst, b, 0644)
                                fallback := filepath.Join(distRoot, "wwwroot", "data", "images", rel)
                                if strings.Contains(pair[1], "images") {
                                        fallback = filepath.Join(distRoot, "wwwroot", "data", "images", rel)
                                } else {
                                        fallback = filepath.Join(distRoot, "wwwroot", "data", "sounds", rel)
                                }
                                os.MkdirAll(filepath.Dir(fallback), 0755)
                                os.WriteFile(fallback, b, 0644)
                                return nil
                        })
                }
        }

        wwwroot := filepath.Join(distRoot, "wwwroot")

        safeName := strings.ReplaceAll(filepath.Base(projectPath), " ", "-")
        timestamp := time.Now().Format("20060102-1504")
        downloads := getDownloadsFolder()
        os.MkdirAll(downloads, 0755)
        finalZip := filepath.Join(downloads, fmt.Sprintf("%s-HOST-%s.zip", safeName, timestamp))
        unpacked := filepath.Join(downloads, fmt.Sprintf("%s-HOST-%s", safeName, timestamp))
        os.RemoveAll(unpacked)

        zf, errZip := os.Create(finalZip)
        if errZip == nil {
                zw := zip.NewWriter(zf)
                filepath.Walk(wwwroot, func(p string, info fs.FileInfo, err error) error {
                        if err != nil || info.IsDir() {
                                return nil
                        }
                        rel, _ := filepath.Rel(wwwroot, p)
                        f, _ := zw.Create(filepath.ToSlash(rel))
                        src, _ := os.Open(p)
                        if src != nil {
                                io.Copy(f, src)
                                src.Close()
                        }
                        dstFile := filepath.Join(unpacked, rel)
                        os.MkdirAll(filepath.Dir(dstFile), 0755)
                        b, _ := os.ReadFile(p)
                        os.WriteFile(dstFile, b, 0644)
                        return nil
                })
                zw.Close()
                zf.Close()
                if a.ctx != nil {
                        runtime.LogInfo(a.ctx, fmt.Sprintf("[COPY] PACZKA: %s (bez odpalania)", finalZip))
                }
        }

        return finalZip, nil
}
