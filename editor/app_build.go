package main

import (
        "archive/zip"
        "fmt"
        "io"
        "io/fs"
        "os"
        "os/exec"
        "path/filepath"
        "strings"
        "time"
)



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

