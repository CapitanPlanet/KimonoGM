package main

import (
    "encoding/base64"
    "fmt"
    "io"
    "io/fs"
    "os"
    "path/filepath"
    "strings"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) SelectAudioFile() (string, error) {
    return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: "Wybierz audio", Filters: []runtime.FileFilter{{DisplayName: "Audio", Pattern: "*.mp3;*.wav;*.ogg"}}})
}
func (a *App) ImportAudioAsset(srcPath string, audioType string) (string, error) {
    if a.projectPath == "" { return "", fmt.Errorf("brak projektu") }
    if strings.TrimSpace(audioType) == "" { audioType = "sfx" }
    sub := strings.ToLower(strings.TrimSpace(audioType))
    if sub != "sfx" && sub != "voice" && sub != "music" { sub = "sfx" }
    dstDir := filepath.Join(a.projectPath, "sounds", sub)
    if err := os.MkdirAll(dstDir, 0755); err != nil { return "", err }
    dst := filepath.Join(dstDir, filepath.Base(srcPath))
    in, err := os.Open(srcPath)
    if err != nil { return "", err }
    defer in.Close()
    out, err := os.Create(dst)
    if err != nil { return "", err }
    defer out.Close()
    if _, err := io.Copy(out, in); err != nil { return "", err }
    return filepath.ToSlash(filepath.Join("sounds", sub, filepath.Base(srcPath))), nil
}
func (a *App) ListAudioAssets(projectPath string) ([]string, error) {
    if projectPath == "" { projectPath = a.projectPath }
    if projectPath == "" { return []string{}, nil }
    var out []string
    baseSounds := filepath.Join(projectPath, "sounds")
    _ = filepath.WalkDir(baseSounds, func(p string, d fs.DirEntry, err error) error {
        if err != nil || d.IsDir() { return nil }
        low := strings.ToLower(d.Name())
        if strings.HasSuffix(low, ".mp3") || strings.HasSuffix(low, ".wav") || strings.HasSuffix(low, ".ogg") {
            rel, _ := filepath.Rel(projectPath, p)
            out = append(out, filepath.ToSlash(rel))
        }
        return nil
    })
    return out, nil
}
func (a *App) DeleteAudioAsset(relPath string) error {
    if a.projectPath == "" { return fmt.Errorf("brak projektu") }
    relPath = a.cleanAssetPath(relPath)
    target := filepath.Join(a.projectPath, relPath)
    clean := filepath.Clean(target)
    if !a.isInsideProject(clean) { return fmt.Errorf("delete blocked") }
    return os.Remove(clean)
}
func (a *App) GetAudioBase64(relPath string) (string, error) {
    if a.projectPath == "" { return "", fmt.Errorf("brak projektu") }
    if strings.TrimSpace(relPath) == "" { return "", fmt.Errorf("pusty asset") }
    relPath = a.cleanAssetPath(relPath)
    full := filepath.Join(a.projectPath, relPath)
    full = filepath.Clean(full)
    if _, err := os.Stat(full); os.IsNotExist(err) {
        if found, ok := a.findFileByName(relPath); ok { full = found } else { full = filepath.Join(a.projectPath, "sounds", filepath.Base(relPath)) }
    }
    b, err := os.ReadFile(full)
    if err != nil { return "", err }
    ext := strings.ToLower(filepath.Ext(full))
    mime := "audio/mpeg"
    switch ext {
    case ".wav": mime = "audio/wav"
    case ".ogg": mime = "audio/ogg"
    }
    return fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(b)), nil
}
