package main

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

func (a *App) ReadJSON(fullPath string) (string, error) {
    b, err := os.ReadFile(filepath.Clean(fullPath))
    return string(b), err
}
func (a *App) WriteJSON(fullPath string, content string) error {
    clean := filepath.Clean(fullPath)
    if err := os.MkdirAll(filepath.Dir(clean), 0755); err != nil { return err }
    return os.WriteFile(clean, []byte(content), 0644)
}
func (a *App) SaveJsonFile(filename string, content string) error {
    if a.projectPath == "" { return fmt.Errorf("brak projectPath") }
    base := filepath.Base(filename)
    target := filepath.Join(a.projectPath, "Data", base)
    safe, err := a.safeJoin(target)
    if err != nil { return err }
    if err := os.MkdirAll(filepath.Dir(safe), 0755); err != nil { return err }
    return os.WriteFile(safe, []byte(content), 0644)
}
func (a *App) ListFiles(dirPath string, ext string) ([]string, error) {
    clean := filepath.Clean(dirPath)
    entries, err := os.ReadDir(clean)
    if err != nil { return []string{}, nil }
    var out []string
    for _, e := range entries {
        if !e.IsDir() && (ext == "" || strings.HasSuffix(strings.ToLower(e.Name()), strings.ToLower(ext))) {
            out = append(out, e.Name())
        }
    }
    return out, nil
}
func (a *App) DeleteFile(projectPath string, relativePath string) error {
    if projectPath == "" { projectPath = a.projectPath }
    if projectPath == "" { return fmt.Errorf("brak projektu") }
    target := filepath.Join(projectPath, a.cleanAssetPath(relativePath))
    if !a.isInsideProject(filepath.Clean(target)) {
        return fmt.Errorf("delete blocked - outside project")
    }
    return os.Remove(target)
}
