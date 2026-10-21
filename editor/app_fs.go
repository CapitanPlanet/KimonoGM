package main

import (
        "fmt"
        "io"
        "os"
        "path/filepath"
        "strings"
)



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

