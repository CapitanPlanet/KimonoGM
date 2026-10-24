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

func getNextExeName(downloads, safeName string) (string, string) {
	baseExe := filepath.Join(downloads, fmt.Sprintf("%s.exe", safeName))
	if _, err := os.Stat(baseExe); os.IsNotExist(err) {
		return baseExe, ""
	}
	for v := 2; v < 100; v++ {
		candidate := filepath.Join(downloads, fmt.Sprintf("%s_v%d.exe", safeName, v))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate, fmt.Sprintf("_v%d", v)
		}
	}
	return filepath.Join(downloads, fmt.Sprintf("%s_v%d.exe", safeName, time.Now().Unix())), fmt.Sprintf("_v%d", time.Now().Unix())
}

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

_ = a.StopGame() // bezpieczne
	time.Sleep(500 * time.Millisecond)

	cmd := exec.Command("dotnet", "publish", "-c", "Release", "-o", distRoot, "-p:StaticWebAssetsCompressContent=false", "-p:StaticWebAssetCompressionMode=None", "-p:PublishTrimmed=false", "-p:BlazorWebAssemblyEnableTrimming=false")
	cmd.Dir = engineRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("publish failed: %w", err)
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

	downloads := getDownloadsFolder()
	gameName := filepath.Base(a.projectPath)
	safeName := strings.ReplaceAll(gameName, " ", "-")
	finalExe, _ := getNextExeName(downloads, safeName)

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

	var launcherDir string
	if launcherProj == "" {
		launcherDir = filepath.Join(filepath.Dir(engineRoot), "Launcher")
		_ = os.MkdirAll(launcherDir, 0755)
		launcherProj = filepath.Join(launcherDir, "Launcher.csproj")
	} else {
		launcherDir = filepath.Dir(launcherProj)
	}
	_ = os.MkdirAll(launcherDir, 0755)
	_ = os.Remove(filepath.Join(launcherDir, "Program.cs"))

	for _, ic := range []string{
		filepath.Join(filepath.Dir(engineRoot), "..", "editor", "build", "windows", "icon.ico"),
		filepath.Join(engineRoot, "..", "editor", "build", "windows", "icon.ico"),
		"editor/build/windows/icon.ico",
		filepath.Join("editor", "build", "windows", "icon.ico"),
	} {
		if _, err := os.Stat(ic); err == nil {
			b, _ := os.ReadFile(ic)
			_ = os.WriteFile(filepath.Join(launcherDir, "icon.ico"), b, 0644)
			break
		}
	}

	// FIX v5: csproj MUSI zawierac Content Include dla wwwroot.zip i icon.ico
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
    <IncludeAllContentForSelfExtract>true</IncludeAllContentForSelfExtract>
    <ApplicationIcon>icon.ico</ApplicationIcon>
    <Win32Icon>icon.ico</Win32Icon>
  </PropertyGroup>
  <ItemGroup>
    <PackageReference Include="Microsoft.Web.WebView2" Version="1.0.2739.15" />
  </ItemGroup>
  <ItemGroup>
    <Content Include="icon.ico" Condition="Exists('icon.ico')">
      <CopyToOutputDirectory>PreserveNewest</CopyToOutputDirectory>
    </Content>
    <Content Include="wwwroot.zip" Condition="Exists('wwwroot.zip')">
      <CopyToOutputDirectory>PreserveNewest</CopyToOutputDirectory>
      <CopyToPublishDirectory>PreserveNewest</CopyToPublishDirectory>
    </Content>
  </ItemGroup>
</Project>`
	_ = os.WriteFile(launcherProj, []byte(csprojContent), 0644)

	programCs := `using System;
using System.IO;
using System.IO.Compression;
using System.Linq;
using System.Net;
using System.Net.Sockets;
using System.Threading;
using System.Windows.Forms;
using Microsoft.Web.WebView2.WinForms;
using Microsoft.Web.WebView2.Core;

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

    static string FindWwwrootZip() {
        var candidates = new[] {
            Path.Combine(AppContext.BaseDirectory, "wwwroot.zip"),
            Path.Combine(Path.GetDirectoryName(Application.ExecutablePath), "wwwroot.zip"),
            Path.Combine(Directory.GetCurrentDirectory(), "wwwroot.zip"),
            Path.Combine(Path.GetTempPath(), "KimonoGM", "wwwroot.zip"),
        };
        // w single file mode AppContext.BaseDirectory to temp .net folder
        foreach (var p in candidates) if (File.Exists(p)) return p;
        // szukaj rekurencyjnie w baseDir
        try {
            foreach (var f in Directory.GetFiles(AppContext.BaseDirectory, "wwwroot.zip", SearchOption.AllDirectories)) {
                if (File.Exists(f)) return f;
            }
        } catch {}
        return null;
    }

    [STAThread]
    static void Main() {
        Application.EnableVisualStyles();
        Application.SetCompatibleTextRenderingDefault(false);

        var exeName = Path.GetFileNameWithoutExtension(Application.ExecutablePath);
        var tempRoot = Path.Combine(Path.GetTempPath(), "KimonoGM", exeName);
        var wwwroot = Path.Combine(tempRoot, "wwwroot");

        // znajdz zip
        var wwwrootZip = FindWwwrootZip();

        try {
            if (!string.IsNullOrEmpty(wwwrootZip) && File.Exists(wwwrootZip)) {
                var marker = Path.Combine(tempRoot, "_ver.txt");
                var zipTime = File.GetLastWriteTimeUtc(wwwrootZip).Ticks.ToString() + "_" + new FileInfo(wwwrootZip).Length.ToString();
                var needExtract = true;
                if (File.Exists(marker) && File.ReadAllText(marker) == zipTime && Directory.Exists(wwwroot) && File.Exists(Path.Combine(wwwroot, "index.html"))) {
                    needExtract = false;
                }
                if (needExtract) {
                    if (Directory.Exists(wwwroot)) try { Directory.Delete(wwwroot, true); } catch {}
                    Directory.CreateDirectory(wwwroot);
                    ZipFile.ExtractToDirectory(wwwrootZip, wwwroot, true);
                    Directory.CreateDirectory(tempRoot);
                    File.WriteAllText(marker, zipTime);
                }
            }
        } catch (Exception ex) {
            MessageBox.Show("Blad wypakowania wwwroot.zip: " + ex.Message + "\nZip: " + wwwrootZip + "\nTemp: " + tempRoot);
        }

        if (!File.Exists(Path.Combine(wwwroot, "index.html"))) {
            var fallback = new[] {
                Path.Combine(AppContext.BaseDirectory, "wwwroot"),
                Path.Combine(Path.GetDirectoryName(Application.ExecutablePath), "wwwroot"),
                Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), "Downloads", "wwwroot"),
            };
            foreach (var c in fallback) {
                if (File.Exists(Path.Combine(c, "index.html"))) { wwwroot = c; break; }
            }
        }

        if (!File.Exists(Path.Combine(wwwroot, "index.html"))) {
            MessageBox.Show("Brak wwwroot/index.html\nSzukano w:\n" + wwwroot + "\nZip: " + (wwwrootZip ?? "nie znaleziono") + "\nBaseDir: " + AppContext.BaseDirectory + "\nExeDir: " + Path.GetDirectoryName(Application.ExecutablePath));
            return;
        }

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
                    var resp = ctx.Response;
                    var urlPath = ctx.Request.Url.AbsolutePath.TrimStart('/');
                    if (string.IsNullOrEmpty(urlPath)) urlPath = "index.html";
                    var filePath = Path.GetFullPath(Path.Combine(wwwroot, urlPath.Replace('/', Path.DirectorySeparatorChar)));
                    if (!filePath.StartsWith(Path.GetFullPath(wwwroot), StringComparison.OrdinalIgnoreCase)) {
                        filePath = Path.Combine(wwwroot, "index.html");
                    }
                    if (!File.Exists(filePath) && !urlPath.Contains(".")) {
                        filePath = Path.Combine(wwwroot, "index.html");
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

        var baseDir = AppContext.BaseDirectory;
        var iconPath = Path.Combine(baseDir, "icon.ico");
        if (!File.Exists(iconPath)) iconPath = Path.Combine(Path.GetDirectoryName(Application.ExecutablePath), "icon.ico");
        System.Drawing.Icon icon = null;
        if (File.Exists(iconPath)) { try { icon = new System.Drawing.Icon(iconPath); } catch {} }

        var form = new Form() { Width=1280, Height=800, Text="KimonoGM Game", StartPosition=FormStartPosition.CenterScreen, Icon = icon };
        var webView = new WebView2() { Dock=DockStyle.Fill };
        form.Controls.Add(webView);
        form.FormClosed += (s,e) => { try { http.Stop(); } catch {} Environment.Exit(0); };
        form.Load += async (s,e) => {
            try {
                var webView2Temp = Path.Combine(Path.GetTempPath(), "KimonoGM", "WebView2", exeName);
                Directory.CreateDirectory(webView2Temp);
                var env = await CoreWebView2Environment.CreateAsync(null, webView2Temp);
                await webView.EnsureCoreWebView2Async(env);
            } catch {
                await webView.EnsureCoreWebView2Async();
            }
            webView.CoreWebView2.Navigate($"http://localhost:{port}/");
        };
        Application.Run(form);
    }
}`
	_ = os.WriteFile(filepath.Join(launcherDir, "Program.cs"), []byte(programCs), 0644)

	// Przygotuj wwwroot.zip dla Launchera - MUSI istniec przed publish
	srcWwwroot := filepath.Join(distRoot, "wwwroot")
	zipPath := filepath.Join(launcherDir, "wwwroot.zip")
	_ = os.Remove(zipPath)
	zf, err := os.Create(zipPath)
	if err != nil {
		return "", err
	}
	zw := zip.NewWriter(zf)
	filepath.Walk(srcWwwroot, func(p string, info fs.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(srcWwwroot, p)
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

	tmpOut := filepath.Join(os.TempDir(), "kimono_launcher_publish")
	os.RemoveAll(tmpOut)

	c := exec.Command("dotnet", "publish", launcherProj, "-c", "Release", "-r", "win-x64", "--self-contained", "true", "-p:PublishSingleFile=true", "-p:IncludeNativeLibrariesForSelfExtract=true", "-p:IncludeAllContentForSelfExtract=true", "-o", tmpOut)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return "", fmt.Errorf("launcher publish failed: %w", err)
	}

	var exePath string
	if _, err := os.Stat(filepath.Join(tmpOut, "Launcher.exe")); err == nil {
		exePath = filepath.Join(tmpOut, "Launcher.exe")
	} else {
		if files, _ := os.ReadDir(tmpOut); files != nil {
			for _, f := range files {
				if strings.HasSuffix(strings.ToLower(f.Name()), ".exe") {
					exePath = filepath.Join(tmpOut, f.Name())
					break
				}
			}
		}
	}
	if exePath == "" {
		return "", fmt.Errorf("nie znaleziono Launcher.exe w %s", tmpOut)
	}

	// Sprawdz czy wwwroot.zip jest w tmpOut - jesli nie, skopiuj recznie i zbuduj exe z embedded zip via /resource?
	// Dla pewnosci skopiujemy zip obok finalnego exe jako fallback i dodamy go jako embedded resource do exe przez dotnet publish
	// Ale teraz publish juz ma wwwroot.zip jako Content, wiec powinien byc w tmpOut

	b, _ := os.ReadFile(exePath)
	if err := os.WriteFile(finalExe, b, 0644); err != nil {
		return "", err
	}

	// Posprzataj stary syf
	_ = os.RemoveAll(filepath.Join(downloads, "wwwroot"))
	if files, err := os.ReadDir(downloads); err == nil {
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".WebView2") || strings.Contains(f.Name(), ".exe.WebView2") {
				_ = os.RemoveAll(filepath.Join(downloads, f.Name()))
			}
			if strings.HasSuffix(f.Name(), ".zip") && strings.Contains(f.Name(), "hthrtg") {
				_ = os.Remove(filepath.Join(downloads, f.Name()))
			}
		}
	}

	exec.Command("explorer", "/select,", finalExe).Start()
	return finalExe, nil
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
	cmd := exec.Command("dotnet", "publish", "-c", "Release", "-o", distRoot, "-p:StaticWebAssetsCompressContent=false", "-p:StaticWebAssetCompressionMode=None", "-p:PublishTrimmed=false", "-p:BlazorWebAssemblyEnableTrimming=false")
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
		f, _ := zw.Create(relSlash)
		src, _ := os.Open(p)
		if src != nil {
			io.Copy(f, src)
			src.Close()
		}
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

func (a *App) ExportPack(projectPath string) (string, error) {
	return a.BuildPublicPack(projectPath)
}
