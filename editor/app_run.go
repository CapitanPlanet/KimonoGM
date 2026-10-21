package main

import (
        "fmt"
        "io/fs"
        "os"
        "os/exec"
        "path/filepath"
        "strings"
        "time"

        "github.com/wailsapp/wails/v2/pkg/runtime"
)



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



