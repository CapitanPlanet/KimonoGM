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

	fmt.Println("[BUILDER] Czyszcze obj/bin/wwwroot/data...")
	os.RemoveAll(filepath.Join(engineRoot, "obj"))
	os.RemoveAll(filepath.Join(engineRoot, "bin"))
	os.RemoveAll(filepath.Join(engineRoot, "wwwroot", "data"))
	// usun tez stare backupy ktore zostaly
	os.RemoveAll(filepath.Join(engineRoot, "wwwroot", "data_backup_before_build"))
	os.RemoveAll(filepath.Join(engineRoot, "Assets_backup_before_build"))
	os.RemoveAll(filepath.Join(engineRoot, "MakerEngine.csproj_backup_before_build"))

	fmt.Println("[BUILDER] 1/3 dotnet publish...")
	publishDir := filepath.Join(os.TempDir(), "kimono_publish")
	os.RemoveAll(publishDir)
	// FIX: dodaj /p:ErrorOnDuplicatePublishOutputFiles=false i czyscic static web assets
	cmd := exec.Command("dotnet", "publish", engineProj, "-c", "Release", "-o", publishDir, fmt.Sprintf("-p:GameAssetsRoot=%s", *gamePath), "/p:StaticWebAssetsEnabled=false")
	// najpierw proba bez StaticWebAssets
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		// jesli fail z StaticWebAssetsEnabled=false, sprobuj normalnie ale po czyszczeniu
		fmt.Println("[BUILDER] Retry z normalnym publish po czyszczeniu...")
		os.RemoveAll(filepath.Join(engineRoot, "obj"))
		cmd2 := exec.Command("dotnet", "publish", engineProj, "-c", "Release", "-o", publishDir, fmt.Sprintf("-p:GameAssetsRoot=%s", *gamePath))
		cmd2.Stdout = os.Stdout
		cmd2.Stderr = os.Stderr
		if err2 := cmd2.Run(); err2 != nil {
			// jesli dalej blad o day2.json, to znaczy ze w csproj jest hardcoded asset
			// stworz pusty day2.json tymczasowo zeby publish przeszedl, potem go usuniemy
			fmt.Println("[BUILDER] FIX: tworze tymczasowy day2.json zeby obejsc StaticWebAssets...")
			tmpDataDir := filepath.Join(engineRoot, "wwwroot", "data")
			os.MkdirAll(tmpDataDir, 0755)
			os.WriteFile(filepath.Join(tmpDataDir, "day2.json"), []byte("{}"), 0644)
			os.RemoveAll(filepath.Join(engineRoot, "obj"))
			cmd3 := exec.Command("dotnet", "publish", engineProj, "-c", "Release", "-o", publishDir, fmt.Sprintf("-p:GameAssetsRoot=%s", *gamePath))
			cmd3.Stdout = os.Stdout
			cmd3.Stderr = os.Stderr
			if err3 := cmd3.Run(); err3 != nil {
				fmt.Printf("[BUILDER] FAILED: %v\n", err3)
				os.Exit(1)
			}
			os.Remove(filepath.Join(tmpDataDir, "day2.json"))
		}
	}

	pubData := filepath.Join(publishDir, "wwwroot", "data")
	if _, err := os.Stat(pubData); os.IsNotExist(err) {
		os.MkdirAll(pubData, 0755)
		copyDir(sourceData, pubData)
	}

	fmt.Println("[BUILDER] 2/3 kopiowanie wwwroot...")
	wwwrootSrc := filepath.Join(publishDir, "wwwroot")
	wwwrootDst := filepath.Join(filepath.Dir(*outPath), "wwwroot")
	os.RemoveAll(wwwrootDst)
	copyDir(wwwrootSrc, wwwrootDst)

	finalData := filepath.Join(wwwrootDst, "data")
	os.MkdirAll(finalData, 0755)
	// wyczysc co publish mogl skopiowac z engine (day2.json)
	os.RemoveAll(finalData)
	os.MkdirAll(finalData, 0755)
	copyDir(sourceData, finalData)

	// project.janproj fix
	possibleJanproj := []string{
		filepath.Join(*gamePath, "project.janproj"),
		filepath.Join(*gamePath, "Data", "project.janproj"),
		filepath.Join(sourceData, "project.janproj"),
	}
	for _, src := range possibleJanproj {
		if _, err := os.Stat(src); err == nil {
			copyFile(src, filepath.Join(wwwrootDst, "project.janproj"))
			copyFile(src, filepath.Join(finalData, "project.janproj"))
			fmt.Printf("  - project.janproj z %s\n", src)
			break
		}
	}

	// images fix
	possibleImages := []string{
		filepath.Join(*gamePath, "images"),
		filepath.Join(*gamePath, "Images"),
		filepath.Join(*gamePath, "Data", "images"),
		filepath.Join(sourceData, "images"),
	}
	for _, src := range possibleImages {
		if st, err := os.Stat(src); err == nil && st.IsDir() {
			copyDir(src, filepath.Join(wwwrootDst, "images"))
			copyDir(src, filepath.Join(wwwrootDst, "Images"))
			fmt.Printf("  - images z %s\n", src)
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
			copyFile(p, filepath.Join(wwwrootDst, "images", rel))
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
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	fmt.Printf("KimonoGM - serwuje %s na http://localhost:%d\n", wwwroot, port)
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
		fmt.Printf("[BUILDER] Launcher fail: %v\n", err)
	} else {
		fmt.Printf("[BUILDER] GOTOWE: %s\n", *outPath)
	}
	// final cleanup
	os.RemoveAll(filepath.Join(engineRoot, "wwwroot", "data"))
}
