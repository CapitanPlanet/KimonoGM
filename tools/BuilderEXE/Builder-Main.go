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
func killStale() {
	if runtime.GOOS != "windows" {
		return
	}
	exec.Command("taskkill", "/F", "/IM", "qqppppwww.exe").Run()
	for i := 2; i <= 50; i++ {
		exec.Command("taskkill", "/F", "/IM", fmt.Sprintf("qqppppwww_v%d.exe", i)).Run()
	}
	exec.Command("dotnet", "build-server", "shutdown").Run()
	time.Sleep(400 * time.Millisecond)
}
func findEngineRoot(builderDir string) string {
	cands := []string{
		filepath.Join(filepath.Dir(filepath.Dir(builderDir)), "engine"),
		`C:\Users\MD-Core\KimonoGM\engine`,
		filepath.Join(builderDir, "..", "..", "engine"),
	}
	for _, c := range cands {
		if _, err := os.Stat(filepath.Join(c, "MakerEngine.csproj")); err == nil {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	return cands[0]
}

func main() {
	gamePath := flag.String("game", "", "path to game")
	outPath := flag.String("out", "", "output")
	flag.Parse()
	if *gamePath == "" {
		fmt.Println("Uzycie: --game path [--out exe]")
		os.Exit(1)
	}
	finalOut := *outPath
	if finalOut == "" {
		gn := filepath.Base(*gamePath)
		dl := filepath.Join(os.Getenv("USERPROFILE"), "Downloads")
		finalOut = filepath.Join(dl, gn+".exe")
		if _, err := os.Stat(finalOut); err == nil {
			for i := 2; i < 50; i++ {
				t := filepath.Join(dl, fmt.Sprintf("%s_v%d.exe", gn, i))
				if _, err := os.Stat(t); os.IsNotExist(err) {
					finalOut = t
					break
				}
			}
		}
	}
	fmt.Println("[BUILDER v41 FIX CSS]")
	killStale()
	cwd, _ := os.Getwd()
	engineRoot := findEngineRoot(cwd)
	engineProj := filepath.Join(engineRoot, "MakerEngine.csproj")
	engineWwwrootData := filepath.Join(engineRoot, "wwwroot", "data")
	engineAssets := filepath.Join(engineRoot, "wwwroot", "images")
	builderDir := cwd

	sourceData := ""
	for _, c := range []string{filepath.Join(*gamePath, "Data"), *gamePath} {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			// check if has json
			has := false
			filepath.Walk(c, func(p string, info os.FileInfo, e error) error {
				if strings.HasSuffix(strings.ToLower(p), ".json") {
					has = true
				}
				return nil
			})
			if has {
				sourceData = c
				break
			}
		}
	}
	if sourceData == "" {
		sourceData = *gamePath
	}

	// backup wwwroot/data + images
	wb := engineWwwrootData + "_backup"
	ab := engineAssets + "_backup"
	we, ae := false, false
	if _, err := os.Stat(engineWwwrootData); err == nil {
		we = true
		os.RemoveAll(wb)
		os.Rename(engineWwwrootData, wb)
	}
	if _, err := os.Stat(engineAssets); err == nil {
		ae = true
		os.RemoveAll(ab)
		os.Rename(engineAssets, ab)
	}
	os.MkdirAll(engineWwwrootData, 0755)
	copyDir(sourceData, engineWwwrootData)
	os.MkdirAll(engineAssets, 0755)
	filepath.Walk(*gamePath, func(p string, info os.FileInfo, e error) error {
		if e != nil || info == nil || info.IsDir() {
			return nil
		}
		l := strings.ToLower(info.Name())
		if strings.HasSuffix(l, ".webp") || strings.HasSuffix(l, ".png") || strings.HasSuffix(l, ".jpg") {
			copyFile(p, filepath.Join(engineAssets, info.Name()))
		}
		return nil
	})

	defer func() {
		os.RemoveAll(engineWwwrootData)
		if we {
			os.Rename(wb, engineWwwrootData)
		}
		os.RemoveAll(engineAssets)
		if ae {
			os.Rename(ab, engineAssets)
		}
		os.RemoveAll(filepath.Join(engineRoot, "obj"))
		os.RemoveAll(filepath.Join(engineRoot, "bin"))
	}()

	fmt.Println("[1/3] dotnet publish...")
	pubOut := filepath.Join(builderDir, "publish_tmp")
	os.RemoveAll(pubOut)
	cmdPub := exec.Command("dotnet", "publish", engineProj, "-c", "Release", "-o", pubOut, "-p:StaticWebAssetsCompressContent=false", "-p:StaticWebAssetCompressionMode=None")
	cmdPub.Stdout = os.Stdout
	cmdPub.Stderr = os.Stderr
	if err := cmdPub.Run(); err != nil {
		os.Exit(1)
	}

	embedRoot := filepath.Join(builderDir, "embed_tmp")
	os.RemoveAll(embedRoot)
	os.MkdirAll(embedRoot, 0755)
	embedWwwroot := filepath.Join(embedRoot, "wwwroot")
	copyDir(filepath.Join(pubOut, "wwwroot"), embedWwwroot)

	// CLEAN: only data/ and images/ no duplicates
	dataDst := filepath.Join(embedWwwroot, "data")
	imgDst := filepath.Join(embedWwwroot, "images")
	os.MkdirAll(dataDst, 0755)
	os.MkdirAll(imgDst, 0755)

	// json -> data/
	filepath.Walk(sourceData, func(p string, info os.FileInfo, e error) error {
		if e != nil || info == nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(info.Name()), ".json") {
			copyFile(p, filepath.Join(dataDst, info.Name()))
		}
		return nil
	})
	// project.janproj -> root + data/
	for _, s := range []string{filepath.Join(*gamePath, "project.janproj"), filepath.Join(sourceData, "project.janproj")} {
		if _, err := os.Stat(s); err == nil {
			copyFile(s, filepath.Join(embedWwwroot, "project.janproj"))
			copyFile(s, filepath.Join(dataDst, "project.janproj"))
			break
		}
	}
	// images -> images/ only
	filepath.Walk(*gamePath, func(p string, info os.FileInfo, e error) error {
		if e != nil || info == nil || info.IsDir() {
			return nil
		}
		l := strings.ToLower(info.Name())
		if strings.HasSuffix(l, ".webp") || strings.HasSuffix(l, ".png") || strings.HasSuffix(l, ".jpg") {
			copyFile(p, filepath.Join(imgDst, info.Name()))
		}
		return nil
	})

	fmt.Println("[2/3] ikona...")
	iconSrc := filepath.Join(builderDir, "icon.ico")
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
		syso := filepath.Join(embedRoot, "rsrc.syso")
		exec.Command(rsrcPath, "-ico", iconDst, "-o", syso).Run()
	}

	// v41 LAUNCHER - FIX CSS: no data/ fallback for _framework/css
	launcher := `package main
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
paths:=[]string{
os.Getenv("PROGRAMFILES")+` + "`\\Google\\Chrome\\Application\\chrome.exe`" + `,
os.Getenv("PROGRAMFILES(X86)")+` + "`\\Google\\Chrome\\Application\\chrome.exe`" + `,
os.Getenv("LOCALAPPDATA")+` + "`\\Google\\Chrome\\Application\\chrome.exe`" + `,
os.Getenv("PROGRAMFILES")+` + "`\\Microsoft\\Edge\\Application\\msedge.exe`" + `,
}
f:=""
for _,p:=range paths{
if _,err:=os.Stat(p);err==nil{f=p;break}
}
if f!=""{
exec.Command(f,"--app="+url,"--window-size=1280,800").Start()
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
mime.AddExtensionType(".css","text/css")
listener,_:=net.Listen("tcp","127.0.0.1:0")
port:=listener.Addr().(*net.TCPAddr).Port
listener.Close()
http.HandleFunc("/__quit",func(w http.ResponseWriter,r *http.Request){
w.Header().Set("Access-Control-Allow-Origin","*")
w.Write([]byte("bye"))
go func(){time.Sleep(300*time.Millisecond);os.Exit(0)}()
})
http.HandleFunc("/",func(w http.ResponseWriter,r *http.Request){
if strings.HasPrefix(r.URL.Path,"/__quit"){return}
p:=strings.TrimPrefix(r.URL.Path,"/")
if p==""{p="index.html"}
// FOR _framework and css/js serve directly, no data/ fallback
isAsset:=strings.HasPrefix(p,"_framework/") || strings.HasPrefix(p,"css/") || strings.HasPrefix(p,"js/") || strings.HasSuffix(p,".css") || strings.HasSuffix(p,".js") || strings.HasSuffix(p,".wasm")
var data []byte
var err error
if isAsset{
data,err=fs.ReadFile(wwwrootFS,p)
} else {
 // for game data try p and data/p
 data,err=fs.ReadFile(wwwrootFS,p)
 if err!=nil{
   data,err=fs.ReadFile(wwwrootFS,"data/"+p)
 }
 if err!=nil && !strings.HasPrefix(p,"data/"){
   data,err=fs.ReadFile(wwwrootFS,"data/"+p)
 }
 // also try without data/ prefix if p starts with data/
 if err!=nil && strings.HasPrefix(p,"data/"){
   data,err=fs.ReadFile(wwwrootFS,strings.TrimPrefix(p,"data/"))
 }
}
if err==nil{
 if strings.HasSuffix(p,".html"){
  inject:="<script>function _kq(){try{navigator.sendBeacon('/__quit')}catch(e){}}window.addEventListener('pagehide',_kq);document.addEventListener('visibilitychange',function(){if(document.visibilityState==='hidden')_kq();});</script>"
  html:=string(data)
  if strings.Contains(html,"</body>"){
    html=strings.Replace(html,"</body>",inject+"</body>",1)
  }
  w.Header().Set("Content-Type","text/html; charset=utf-8")
  io.WriteString(w,html)
  return
 }
 if strings.HasSuffix(p,".js"){
  w.Header().Set("Content-Type","application/javascript")
 } else if strings.HasSuffix(p,".css"){
  w.Header().Set("Content-Type","text/css")
 } else if strings.HasSuffix(p,".wasm"){
  w.Header().Set("Content-Type","application/wasm")
 } else if strings.HasSuffix(p,".json"){
  w.Header().Set("Content-Type","application/json")
 } else if strings.HasSuffix(p,".webp"){
  w.Header().Set("Content-Type","image/webp")
 } else if strings.HasSuffix(p,".png"){
  w.Header().Set("Content-Type","image/png")
 }
 w.Write(data)
 return
}
http.FileServer(http.FS(wwwrootFS)).ServeHTTP(w,r)
})
url:=fmt.Sprintf("http://127.0.0.1:%d",port)
fmt.Printf("KimonoGM start %s\n",url)
go openBrowser(url)
log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d",port),nil))
}
`
	os.WriteFile(filepath.Join(embedRoot, "main.go"), []byte(launcher), 0644)
	os.WriteFile(filepath.Join(embedRoot, "go.mod"), []byte("module kimono_launcher\n\ngo 1.21\n"), 0644)

	fmt.Println("[3/3] go build v41...")
	os.MkdirAll(filepath.Dir(finalOut), 0755)
	var bcmd *exec.Cmd
	if runtime.GOOS == "windows" {
		bcmd = exec.Command("go", "build", "-ldflags", "-H=windowsgui", "-o", finalOut, ".")
	} else {
		bcmd = exec.Command("go", "build", "-o", finalOut, ".")
	}
	bcmd.Dir = embedRoot
	bcmd.Stdout = os.Stdout
	bcmd.Stderr = os.Stderr
	if err := bcmd.Run(); err != nil {
		os.Exit(1)
	}
	fmt.Printf("\nGOTOWE v41: %s\n", finalOut)
	if st, err := os.Stat(finalOut); err == nil {
		fmt.Printf("%.1f MB - FIX CSS - _framework/css direct\n", float64(st.Size())/1024/1024)
	}
}
