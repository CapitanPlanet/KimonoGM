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
	"time"
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
		if err != nil || info == nil {
			return nil
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

func taskkillExe(exeName string) {
	if runtime.GOOS != "windows" {
		return
	}
	exec.Command("taskkill", "/F", "/IM", exeName).Run()
}

func main() {
	gamePath := flag.String("game", "", "path to game folder")
	outPath := flag.String("out", "", "output exe path")
	flag.Parse()
	if *gamePath == "" {
		fmt.Println("Uzycie: go run Builder-Main.go --game C:\\path\\to\\game [--out out.exe]")
		os.Exit(1)
	}
	if _, err := os.Stat(*gamePath); err != nil {
		fmt.Printf("BLAD: Gra nie istnieje: %s\n", *gamePath)
		os.Exit(1)
	}

	finalOut := *outPath
	if finalOut == "" {
		gameName := filepath.Base(*gamePath)
		downloads := filepath.Join(os.Getenv("USERPROFILE"), "Downloads")
		finalOut = filepath.Join(downloads, gameName+".exe")
		if _, err := os.Stat(finalOut); err == nil {
			i := 2
			for {
				try := filepath.Join(downloads, fmt.Sprintf("%s_v%d.exe", gameName, i))
				if _, err := os.Stat(try); os.IsNotExist(err) {
					finalOut = try
					break
				}
				i++
			}
		}
	}

	fmt.Printf("[BUILDER] Taskkill: %s\n", filepath.Base(finalOut))
	taskkillExe(filepath.Base(finalOut))
	time.Sleep(500 * time.Millisecond)

	cwd, _ := os.Getwd()
	engineRoot := filepath.Join(filepath.Dir(filepath.Dir(cwd)), "engine")
	engineProj := filepath.Join(engineRoot, "MakerEngine.csproj")
	engineWwwrootData := filepath.Join(engineRoot, "wwwroot", "data")
	engineAssets := filepath.Join(engineRoot, "Assets")
	builderDir := cwd

	fmt.Printf("[BUILDER v35 FINAL STABLE - NO ADMIN NO GCC] Game: %s\n", *gamePath)
	fmt.Printf("[BUILDER] Out: %s\n", finalOut)

	iconSrc := filepath.Join(builderDir, "icon.ico")
	if st, err := os.Stat(iconSrc); err == nil {
		fmt.Printf("[BUILDER] Ikona: %s (%d bytes)\n", iconSrc, st.Size())
	}

	sourceData := ""
	for _, c := range []string{filepath.Join(*gamePath, "Data"), filepath.Join(*gamePath, "data"), *gamePath} {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			hasJson := false
			filepath.Walk(c, func(p string, info os.FileInfo, e error) error {
				if e != nil || info == nil || info.IsDir() {
					return nil
				}
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

	wwwrootBackup := engineWwwrootData + "_backup"
	assetsBackup := engineAssets + "_backup"
	wwwrootExists := false
	assetsExists := false
	if _, err := os.Stat(engineWwwrootData); err == nil {
		wwwrootExists = true
		os.RemoveAll(wwwrootBackup)
		os.Rename(engineWwwrootData, wwwrootBackup)
	}
	if _, err := os.Stat(engineAssets); err == nil {
		assetsExists = true
		os.RemoveAll(assetsBackup)
		os.Rename(engineAssets, assetsBackup)
	}
	os.MkdirAll(engineWwwrootData, 0755)
	copyDir(sourceData, engineWwwrootData)

	gameImagesDst := filepath.Join(engineRoot, "wwwroot", "images")
	os.RemoveAll(gameImagesDst)
	os.MkdirAll(gameImagesDst, 0755)
	filepath.Walk(*gamePath, func(p string, info os.FileInfo, e error) error {
		if e != nil || info == nil || info.IsDir() {
			return nil
		}
		lower := strings.ToLower(info.Name())
		if strings.HasSuffix(lower, ".webp") || strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".jpg") {
			copyFile(p, filepath.Join(gameImagesDst, info.Name()))
		}
		return nil
	})

	defer func() {
		os.RemoveAll(engineWwwrootData)
		if wwwrootExists {
			os.Rename(wwwrootBackup, engineWwwrootData)
		}
		os.RemoveAll(engineAssets)
		if assetsExists {
			os.Rename(assetsBackup, engineAssets)
		}
		os.RemoveAll(filepath.Join(engineRoot, "wwwroot", "images"))
	}()

	fmt.Println("[BUILDER] 1/3 dotnet publish...")
	publishDir := filepath.Join(os.TempDir(), "kimono_publish")
	os.RemoveAll(publishDir)
	cmd := exec.Command("dotnet", "publish", engineProj, "-c", "Release", "-o", publishDir, fmt.Sprintf("-p:GameAssetsRoot=%s", *gamePath))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("dotnet FAILED: %v\n", err)
		os.Exit(1)
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

	finalImagesDst := filepath.Join(embedWwwroot, "images")
	os.MkdirAll(finalImagesDst, 0755)
	filepath.Walk(*gamePath, func(p string, info os.FileInfo, e error) error {
		if e != nil || info == nil || info.IsDir() {
			return nil
		}
		lower := strings.ToLower(info.Name())
		if strings.HasSuffix(lower, ".webp") || strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".jpg") {
			copyFile(p, filepath.Join(finalImagesDst, info.Name()))
			copyFile(p, filepath.Join(finalDataDst, info.Name()))
		}
		return nil
	})

	for _, src := range []string{filepath.Join(*gamePath, "project.janproj")} {
		if _, err := os.Stat(src); err == nil {
			copyFile(src, filepath.Join(embedWwwroot, "project.janproj"))
			break
		}
	}

	fmt.Println("[BUILDER] 2/3 ikona...")
	if _, err := os.Stat(iconSrc); err == nil {
		exec.Command("go", "install", "github.com/akavel/rsrc@latest").Run()
		iconDst := filepath.Join(embedRoot, "icon.ico")
		copyFile(iconSrc, iconDst)
		goBin, _ := exec.Command("go", "env", "GOPATH").Output()
		gopath := strings.TrimSpace(string(goBin))
		if gopath == "" {
			gopath = filepath.Join(os.Getenv("USERPROFILE"), "go")
		}
		rsrcPath := filepath.Join(gopath, "bin", "rsrc.exe")
		if _, err := os.Stat(rsrcPath); err != nil {
			rsrcPath = "rsrc"
		}
		sysoPath := filepath.Join(embedRoot, "rsrc.syso")
		c := exec.Command(rsrcPath, "-ico", iconDst, "-o", sysoPath)
		c.Dir = embedRoot
		c.Run()
	}

	// FINAL STABLE LAUNCHER - bez webview, bez gcc, 100% kompatybilny z userami
	launcherMain := `package main
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
"strings"
"time"
)
//go:embed all:wwwroot
var embeddedFS embed.FS
func openBrowser(url string){
chromePaths:=[]string{
os.Getenv("PROGRAMFILES")+` + "`\\Google\\Chrome\\Application\\chrome.exe`" + `,
os.Getenv("PROGRAMFILES(X86)")+` + "`\\Google\\Chrome\\Application\\chrome.exe`" + `,
os.Getenv("LOCALAPPDATA")+` + "`\\Google\\Chrome\\Application\\chrome.exe`" + `,
os.Getenv("PROGRAMFILES")+` + "`\\Microsoft\\Edge\\Application\\msedge.exe`" + `,
os.Getenv("PROGRAMFILES(X86)")+` + "`\\Microsoft\\Edge\\Application\\msedge.exe`" + `,
}
found:=""
for _,p:=range chromePaths{
if _,err:=os.Stat(p);err==nil{found=p;break}
}
if found!=""{
exec.Command(found,fmt.Sprintf("--app=%s",url),"--window-size=1280,800","--disable-features=Translate").Start()
return
}
exec.Command("cmd","/c","start",url).Start()
}
func main(){
wwwrootFS,_:=fs.Sub(embeddedFS,"wwwroot")
mime.AddExtensionType(".js","application/javascript")
mime.AddExtensionType(".wasm","application/wasm")
mime.AddExtensionType(".json","application/json")
mime.AddExtensionType(".webp","image/webp")
listener,_:=net.Listen("tcp","127.0.0.1:0")
port:=listener.Addr().(*net.TCPAddr).Port
listener.Close()
http.HandleFunc("/__quit",func(w http.ResponseWriter,r *http.Request){
w.Header().Set("Access-Control-Allow-Origin","*")
w.Write([]byte("bye"))
go func(){time.Sleep(300*time.Millisecond);os.Exit(0)}()
})
fsrv:=http.FileServer(http.FS(wwwrootFS))
http.HandleFunc("/",func(w http.ResponseWriter,r *http.Request){
if strings.HasPrefix(r.URL.Path,"/__quit"){return}
path:=strings.TrimPrefix(r.URL.Path,"/")
if path==""{path="index.html"}
if data,err:=fs.ReadFile(wwwrootFS,path);err==nil{
if strings.HasSuffix(path,".html"){
inject:="<script>function q(){try{navigator.sendBeacon('/__quit')}catch(e){}}window.addEventListener('beforeunload',q);window.addEventListener('pagehide',q);window.addEventListener('unload',q);</script>"
html:=string(data)
if strings.Contains(html,"</body>"){
html=strings.Replace(html,"</body>",inject+"</body>",1)
} else {
html+=inject
}
w.Header().Set("Content-Type","text/html; charset=utf-8")
io.WriteString(w,html)
return
}
}
if strings.HasSuffix(r.URL.Path,".js"){
w.Header().Set("Content-Type","application/javascript")
} else if strings.HasSuffix(r.URL.Path,".wasm"){
w.Header().Set("Content-Type","application/wasm")
} else if strings.HasSuffix(r.URL.Path,".json"){
w.Header().Set("Content-Type","application/json")
}
fsrv.ServeHTTP(w,r)
})
url:=fmt.Sprintf("http://127.0.0.1:%d",port)
fmt.Printf("KimonoGM - Gra start http://127.0.0.1:%d\n",port)
go openBrowser(url)
log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d",port),nil))
}
`
	os.WriteFile(filepath.Join(embedRoot, "main.go"), []byte(launcherMain), 0644)
	os.WriteFile(filepath.Join(embedRoot, "go.mod"), []byte("module kimono_launcher\n\ngo 1.21\n"), 0644)

	wwwrootCount := 0
	jsonCount := 0
	filepath.Walk(embedWwwroot, func(p string, info os.FileInfo, e error) error {
		if e != nil || info == nil || info.IsDir() {
			return nil
		}
		wwwrootCount++
		if strings.HasSuffix(strings.ToLower(info.Name()), ".json") {
			jsonCount++
		}
		return nil
	})
	fmt.Printf("[BUILDER] wwwroot: %d plikow (json=%d)\n", wwwrootCount, jsonCount)
	if jsonCount == 0 {
		fmt.Println("[BUILDER] BLAD: Brak json - bedzie 404!")
		os.Exit(1)
	}

	os.MkdirAll(filepath.Dir(finalOut), 0755)
	var buildCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		buildCmd = exec.Command("go", "build", "-ldflags", "-H=windowsgui", "-o", finalOut, ".")
	} else {
		buildCmd = exec.Command("go", "build", "-o", finalOut, ".")
	}
	buildCmd.Dir = embedRoot
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	fmt.Println("[BUILDER] 3/3 go build FINAL...")
	if err := buildCmd.Run(); err != nil {
		fmt.Printf("FAILED: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\nGOTOWE: %s\n", finalOut)
	if st, err := os.Stat(finalOut); err == nil {
		fmt.Printf("Rozmiar: %.1f MB - FINAL STABLE - NO ADMIN NO GCC NO WEBVIEW\n", float64(st.Size())/1024/1024)
	}
	fmt.Println("Userzy uruchomiÄ bez admina, bez Chrome install, bez WebView2 Runtime - wystarczy Edge (jest na kaĹźdym Win10/11)")
}