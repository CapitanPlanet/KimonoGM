package main

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

func (a *App) cleanProjectPath() string {
    if a.projectPath == "" { return "" }
    return filepath.Clean(a.projectPath)
}
func (a *App) isInsideProject(targetPath string) bool {
    if a.projectPath == "" { return false }
    proj := a.cleanProjectPath()
    target := filepath.Clean(targetPath)
    rel, err := filepath.Rel(proj, target)
    if err != nil { return false }
    return !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && rel != ".."
}
func (a *App) safeJoin(elem ...string) (string, error) {
    if a.projectPath == "" { return "", fmt.Errorf("brak projektu") }
    joined := filepath.Join(elem...)
    clean := filepath.Clean(joined)
    if !a.isInsideProject(clean) && clean != a.cleanProjectPath() {
        return "", fmt.Errorf("path traversal blocked: %s", clean)
    }
    return clean, nil
}
func (a *App) cleanAssetPath(input string) string {
    p := strings.TrimSpace(input)
    if p == "" { return "" }
    p = strings.ReplaceAll(p, "\\", "/")
    p = filepath.ToSlash(p)
    if filepath.IsAbs(p) || strings.Contains(p, ":") {
        cleanProject := filepath.ToSlash(filepath.Clean(a.projectPath))
        if cleanProject != "" && strings.Contains(p, cleanProject) {
            p = strings.Replace(p, cleanProject, "", 1)
            p = strings.TrimLeft(p, "/")
        } else {
            p = "images/" + filepath.Base(p)
        }
    }
    p = strings.TrimLeft(p, "/")
    return filepath.FromSlash(p)
}
