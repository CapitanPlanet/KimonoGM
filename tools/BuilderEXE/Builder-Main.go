package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

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

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if rel == "." {
			return nil
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(p, target)
	})
}

func main() {
	gamePath := flag.String("game", "", "path to game folder")
	outPath := flag.String("out", "", "output exe path")
	flag.Parse()
	if *gamePath == "" || *outPath == "" {
		fmt.Println("go run Builder-Main.go --game \"...\" --out \"...\"")
		os.Exit(1)
	}

	cwd, _ := os.Getwd()
	engineRoot := filepath.Join(filepath.Dir(filepath.Dir(cwd)), "engine")
	engineProj := filepath.Join(engineRoot, "MakerEngine.csproj")

	fmt.Printf("[BUILDER] Game: %s\n", *gamePath)

	sourceData := ""
	for _, c := range []string{filepath.Join(*gamePath, "Data"), filepath.Join(*gamePath, "data"), *gamePath} {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			hasJson := false
			filepath.Walk(c, func(p string, info os.FileInfo, e error) error {
				if strings.HasSuffix(strings.ToLower(info.Name()), ".json") {
					hasJson = true
				}
				return nil
			})
			if hasJson {
				sourceData = c
				break
			}
		}
	}
	if sourceData == "" {
		sourceData = *gamePath
	}

	fmt.Println("[BUILDER] Czyszcze obj/bin...")
	os.RemoveAll(filepath.Join(engineRoot, "obj"))
	os.RemoveAll(filepath.Join(engineRoot, "bin"))
	os.RemoveAll(filepath.Join(engineRoot, "wwwroot", "data"))

	fmt.Println("[BUILDER] 1/3 dotnet publish...")
	publishDir := filepath.Join(os.TempDir(), "kimono_publish")
	os.RemoveAll(publishDir)
	cmd := exec.Command("dotnet", "publish", engineProj, "-c", "Release", "-o", publishDir, fmt.Sprintf("-p:GameAssetsRoot=%s", *gamePath))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.RemoveAll(filepath.Join(engineRoot, "obj"))
		tmpDataDir := filepath.Join(engineRoot, "wwwroot", "data")
		os.MkdirAll(tmpDataDir, 0755)
		os.WriteFile(filepath.Join(tmpDataDir, "day2.json"), []byte("{}"), 0644)
		cmd2 := exec.Command("dotnet", "publish", engineProj, "-c", "Release", "-o", publishDir, fmt.Sprintf("-p:GameAssetsRoot=%s", *gamePath))
		cmd2.Stdout = os.Stdout
		cmd2.Stderr = os.Stderr
		cmd2.Run()
		os.RemoveAll(tmpDataDir)
	}

	wwwrootSrc := filepath.Join(publishDir, "wwwroot")
	embedRoot := filepath.Join(os.TempDir(), "kimono_embed")
	os.RemoveAll(embedRoot)
	os.MkdirAll(embedRoot, 0755)
	embedWwwroot := filepath.Join(embedRoot, "wwwroot")
	copyDir(wwwrootSrc, embedWwwroot)

	finalDataDst := filepath.Join(embedWwwroot, "data")
	os.RemoveAll(finalDataDst)
	os.MkdirAll(finalDataDst, 0755)
	copyDir(sourceData, finalDataDst)

	for _, src := range []string{filepath.Join(*gamePath, "project.janproj"), filepath.Join(*gamePath, "Data", "project.janproj"), filepath.Join(sourceData, "project.janproj")} {
		if _, err := os.Stat(src); err == nil {
			copyFile(src, filepath.Join(embedWwwroot, "project.janproj"))
			copyFile(src, filepath.Join(finalDataDst, "project.janproj"))
			break
		}
	}
	for _, src := range []string{filepath.Join(*gamePath, "images"), filepath.Join(*gamePath, "Data", "images"), filepath.Join(sourceData, "images")} {
		if st, err := os.Stat(src); err == nil && st.IsDir() {
			copyDir(src, filepath.Join(embedWwwroot, "images"))
			break
		}
	}
	filepath.Walk(sourceData, func(p string, info os.FileInfo, e error) error {
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(p))
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" {
			rel, _ := filepath.Rel(sourceData, p)
			if !strings.Contains(strings.ToLower(rel), "images") {
				copyFile(p, filepath.Join(embedWwwroot, "images", filepath.Base(p)))
			}
		}
		return nil
	})

	fmt.Println("[BUILDER] 3/3 budowanie exe bez konsoli (windowsgui)...")
	launcherCode := `package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

//go:embed all:wwwroot
var embeddedFS embed.FS

func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		chromePaths := []string{
			os.Getenv("PROGRAMFILES") + "\\Google\\Chrome\\Application\\chrome.exe",
			os.Getenv("PROGRAMFILES(X86)") + "\\Google\\Chrome\\Application\\chrome.exe",
			os.Getenv("LOCALAPPDATA") + "\\Google\\Chrome\\Application\\chrome.exe",
			os.Getenv("PROGRAMFILES") + "\\Microsoft\\Edge\\Application\\msedge.exe",
		}
		found := ""
		for _, p := range chromePaths {
			if _, err := os.Stat(p); err == nil {
				found = p
				break
			}
		}
		if found != "" {
			cmd = found
			args = []string{fmt.Sprintf("--app=%s", url), "--new-window", "--window-size=1280,800"}
			exec.Command(cmd, args...).Start()
			return
		}
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}
	exec.Command(cmd, args...).Start()
}

func main() {
	wwwrootFS, _ := fs.Sub(embeddedFS, "wwwroot")
	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".wasm", "application/wasm")
	mime.AddExtensionType(".dll", "application/octet-stream")
	mime.AddExtensionType(".json", "application/json")

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	http.HandleFunc("/__quit", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("bye"))
		go func() { time.Sleep(300 * time.Millisecond); os.Exit(0) }()
	})

	fsrv := http.FileServer(http.FS(wwwrootFS))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/__quit") {
			return
		}
		if r.URL.Path == "/" || r.URL.Path == "/index.html" || strings.HasSuffix(r.URL.Path, ".html") {
			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				path = "index.html"
			}
			if data, err := fs.ReadFile(wwwrootFS, path); err == nil {
				inject := "<script>(function(){function q(){try{navigator.sendBeacon('/__quit')}catch(e){fetch('/__quit',{method:'POST',keepalive:true})}} window.addEventListener('beforeunload',q); window.addEventListener('pagehide',q);})();</script>"
				html := string(data)
				if strings.Contains(html, "</body>") {
					html = strings.Replace(html, "</body>", inject+"</body>", 1)
				} else {
					html += inject
				}
				w.Header().Set("Content-Type", "text/html")
				io.WriteString(w, html)
				return
			}
		}
		if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		} else if strings.HasSuffix(r.URL.Path, ".wasm") {
			w.Header().Set("Content-Type", "application/wasm")
		}
		fsrv.ServeHTTP(w, r)
	})

	url := fmt.Sprintf("http://localhost:%d", port)
	go openBrowser(url)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil))
}
`
	launcherPath := filepath.Join(embedRoot, "launcher.go")
	os.WriteFile(launcherPath, []byte(launcherCode), 0644)

	os.MkdirAll(filepath.Dir(*outPath), 0755)
	
	var buildCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		buildCmd = exec.Command("go", "build", "-ldflags", "-H=windowsgui", "-o", *outPath, launcherPath)
	} else {
		buildCmd = exec.Command("go", "build", "-o", *outPath, launcherPath)
	}
	buildCmd.Dir = embedRoot
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Printf("[BUILDER] FAILED: %v\n", err)
		os.Exit(1)
	}

	os.RemoveAll(filepath.Join(engineRoot, "wwwroot", "data"))
	fmt.Printf("[BUILDER] GOTOWE (bez konsoli): %s\n", *outPath)
	if st, err := os.Stat(*outPath); err == nil {
		fmt.Printf("Rozmiar: %.1f MB - single file, no cmd window, no collision\n", float64(st.Size())/1024/1024)
	}
}
