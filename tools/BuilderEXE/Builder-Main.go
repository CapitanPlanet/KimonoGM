package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
	engineWwwrootData := filepath.Join(engineRoot, "wwwroot", "data")

	fmt.Printf("[BUILDER] Engine: %s\n", engineProj)
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
	fmt.Printf("[BUILDER] Zrodlo: %s\n", sourceData)

	fmt.Println("[BUILDER] Czyszcze obj/bin...")
	os.RemoveAll(filepath.Join(engineRoot, "obj"))
	os.RemoveAll(filepath.Join(engineRoot, "bin"))
	os.Remove(filepath.Join(engineWwwrootData, ".keep"))

	fmt.Println("[BUILDER] 1/3 dotnet publish...")
	publishDir := filepath.Join(os.TempDir(), "kimono_publish")
	os.RemoveAll(publishDir)
	cmd := exec.Command("dotnet", "publish", engineProj, "-c", "Release", "-o", publishDir, fmt.Sprintf("-p:GameAssetsRoot=%s", *gamePath))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("[BUILDER] FAILED: %v\n", err)
		os.Exit(1)
	}

	pubData := filepath.Join(publishDir, "wwwroot", "data")
	if _, err := os.Stat(pubData); os.IsNotExist(err) {
		os.MkdirAll(pubData, 0755)
		copyDir(sourceData, pubData)
	} else {
		empty := true
		filepath.Walk(pubData, func(p string, info os.FileInfo, e error) error {
			if !info.IsDir() {
				empty = false
			}
			return nil
		})
		if empty {
			copyDir(sourceData, pubData)
		}
	}

	fmt.Println("[BUILDER] 2/3 kopiowanie wwwroot...")
	wwwrootSrc := filepath.Join(publishDir, "wwwroot")
	wwwrootDst := filepath.Join(filepath.Dir(*outPath), "wwwroot")
	os.RemoveAll(wwwrootDst)
	copyDir(wwwrootSrc, wwwrootDst)

	finalData := filepath.Join(wwwrootDst, "data")
	os.MkdirAll(finalData, 0755)
	copyDir(sourceData, finalData)

	// FIX: project.janproj - kopiuj wszedzie gdzie gra go szuka
	fmt.Println("[BUILDER] FIX: project.janproj...")
	possibleJanproj := []string{
		filepath.Join(*gamePath, "project.janproj"),
		filepath.Join(*gamePath, "Data", "project.janproj"),
		filepath.Join(*gamePath, "data", "project.janproj"),
		filepath.Join(sourceData, "project.janproj"),
	}
	for _, src := range possibleJanproj {
		if _, err := os.Stat(src); err == nil {
			// kopiuj do wwwroot root (tam gra go szuka - /project.janproj)
			copyFile(src, filepath.Join(wwwrootDst, "project.janproj"))
			// i do wwwroot/data
			copyFile(src, filepath.Join(finalData, "project.janproj"))
			// i do wwwroot root jako scenes.json fallback?
			fmt.Printf("  - znaleziono %s\n", src)
			break
		}
	}

	// FIX: images/sounds - kopiuj z kazdego mozliwego miejsca
	fmt.Println("[BUILDER] FIX: images/sounds/avatar...")
	possibleImages := []string{
		filepath.Join(*gamePath, "images"),
		filepath.Join(*gamePath, "Images"),
		filepath.Join(*gamePath, "Data", "images"),
		filepath.Join(*gamePath, "Data", "Images"),
		filepath.Join(*gamePath, "data", "images"),
		filepath.Join(sourceData, "images"),
		filepath.Join(sourceData, "Images"),
	}
	for _, src := range possibleImages {
		if st, err := os.Stat(src); err == nil && st.IsDir() {
			// kopiuj do obu wariantow - lowercase i oryginal
			dst1 := filepath.Join(wwwrootDst, "images")
			dst2 := filepath.Join(wwwrootDst, "Images")
			copyDir(src, dst1)
			copyDir(src, dst2)
			fmt.Printf("  - images: %s -> %s\n", src, dst1)
			break
		}
	}
	possibleSounds := []string{
		filepath.Join(*gamePath, "sounds"),
		filepath.Join(*gamePath, "Sounds"),
		filepath.Join(*gamePath, "Data", "sounds"),
		filepath.Join(sourceData, "sounds"),
	}
	for _, src := range possibleSounds {
		if st, err := os.Stat(src); err == nil && st.IsDir() {
			dst := filepath.Join(wwwrootDst, "sounds")
			copyDir(src, dst)
			fmt.Printf("  - sounds: %s\n", src)
			break
		}
	}

	// FIX: jesli avatar jest w data/images/avatar.png itp - skopiuj tez bezposrednio pliki json referencje
	// skopiuj wszystkie pliki z sourceData ktore nie sa json (np avatar.png w Data)
	filepath.Walk(sourceData, func(p string, info os.FileInfo, e error) error {
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(p))
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" || ext == ".gif" {
			rel, _ := filepath.Rel(sourceData, p)
			// jesli to obrazek w Data, skopiuj do wwwroot/images i wwwroot/data
			copyFile(p, filepath.Join(wwwrootDst, "images", rel))
			copyFile(p, filepath.Join(wwwrootDst, "Images", rel))
			copyFile(p, filepath.Join(finalData, rel))
		}
		return nil
	})

	fmt.Println("[BUILDER] Finalny wwwroot/data:")
	filepath.Walk(finalData, func(p string, info os.FileInfo, e error) error {
		if !info.IsDir() {
			rel, _ := filepath.Rel(finalData, p)
			fmt.Printf("  - %s\n", rel)
		}
		return nil
	})
	fmt.Println("[BUILDER] Finalny wwwroot root:")
	for _, f := range []string{"project.janproj", "scenes.json"} {
		if _, err := os.Stat(filepath.Join(wwwrootDst, f)); err == nil {
			fmt.Printf("  - %s OK\n", f)
		} else {
			fmt.Printf("  - %s BRAK\n", f)
		}
	}
	if _, err := os.Stat(filepath.Join(wwwrootDst, "images")); err == nil {
		cnt := 0
		filepath.Walk(filepath.Join(wwwrootDst, "images"), func(p string, info os.FileInfo, e error) error {
			if !info.IsDir() {
				cnt++
			}
			return nil
		})
		fmt.Printf("  - images/ %d plikow\n", cnt)
	}

	fmt.Println("[BUILDER] 3/3 budowanie launchera EXE...")
	launcherCode := `
package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
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
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	wwwroot := filepath.Join(exeDir, "wwwroot")

	if _, err := os.Stat(filepath.Join(wwwroot, "index.html")); os.IsNotExist(err) {
		wwwroot = filepath.Join(exeDir, "wwwroot")
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	fmt.Printf("KimonoGM - serwuje %s na http://localhost:%d\n", wwwroot, port)
	fmt.Println("Zamknij to okno zeby zakonczyc gre")

	fs := http.FileServer(http.Dir(wwwroot))
	http.Handle("/", fs)

	url := fmt.Sprintf("http://localhost:%d", port)
	go openBrowser(url)

	log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil))
}
`
	launcherPath := filepath.Join(os.TempDir(), "kimono_launcher.go")
	os.WriteFile(launcherPath, []byte(launcherCode), 0644)

	os.MkdirAll(filepath.Dir(*outPath), 0755)
	buildCmd := exec.Command("go", "build", "-o", *outPath, launcherPath)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Printf("[BUILDER] Launcher fail, ale wwwroot gotowy: %v\n", err)
	} else {
		fmt.Printf("[BUILDER] GOTOWE: %s\n", *outPath)
		fmt.Printf("[BUILDER] + %s\n", wwwrootDst)
	}
}
