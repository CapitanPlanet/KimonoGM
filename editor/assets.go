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

func (a *App) SelectImageFile() (string, error) {
    return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: "Wybierz obraz", Filters: []runtime.FileFilter{{DisplayName: "Images", Pattern: "*.png;*.jpg;*.jpeg;*.webp"}}})
}
func (a *App) ImportAsset(srcPath string, assetType string) (string, error) {
    if a.projectPath == "" { return "", fmt.Errorf("brak projektu") }
    if strings.TrimSpace(assetType) == "" { assetType = "bg" }
    assetType = strings.ToLower(strings.TrimSpace(assetType))
    dstDir := filepath.Join(a.projectPath, "images")
    if err := os.MkdirAll(dstDir, 0755); err != nil { return "", err }
    origBase := filepath.Base(srcPath)
    base := origBase
    lowBase := strings.ToLower(base)
    hasPrefix := strings.HasPrefix(lowBase, "bg_") || strings.HasPrefix(lowBase, "re_") || strings.HasPrefix(lowBase, "av_")
    if !hasPrefix {
        switch assetType {
        case "av", "avatar", "avatars": base = "av_" + origBase
        case "re", "reaction", "reakcja": base = "re_" + origBase
        default: base = "bg_" + origBase
        }
    }
    dst := filepath.Join(dstDir, base)
    if _, err := os.Stat(dst); err == nil {
        ext := filepath.Ext(base)
        name := strings.TrimSuffix(base, ext)
        for i := 1; i < 100; i++ {
            candidate := filepath.Join(dstDir, fmt.Sprintf("%s_%d%s", name, i, ext))
            if _, err := os.Stat(candidate); os.IsNotExist(err) { dst = candidate; base = filepath.Base(candidate); break }
        }
    }
    in, err := os.Open(srcPath)
    if err != nil { return "", err }
    defer in.Close()
    out, err := os.Create(dst)
    if err != nil { return "", err }
    defer out.Close()
    if _, err = io.Copy(out, in); err != nil { return "", err }
    return "images/" + base, nil
}
func (a *App) findFileByName(name string) (string, bool) {
    base := filepath.Base(name)
    if base == "" || base == "." { return "", false }
    candidate := filepath.Join(a.projectPath, "images", base)
    if _, err := os.Stat(candidate); err == nil { return candidate, true }
    var found string
    _ = filepath.WalkDir(filepath.Join(a.projectPath, "images"), func(p string, d fs.DirEntry, err error) error {
        if err == nil && !d.IsDir() && strings.EqualFold(d.Name(), base) { found = p; return io.EOF }
        return nil
    })
    if found != "" { return found, true }
    candidate = filepath.Join(a.projectPath, base)
    if _, err := os.Stat(candidate); err == nil { return candidate, true }
    return "", false
}
func (a *App) GetImageBase64(relPath string) (string, error) {
    if a.projectPath == "" { return "", fmt.Errorf("brak projektu") }
    if strings.TrimSpace(relPath) == "" { return "", fmt.Errorf("pusty asset") }
    relPath = a.cleanAssetPath(relPath)
    full := filepath.Join(a.projectPath, relPath)
    full = filepath.Clean(full)
    if _, err := os.Stat(full); os.IsNotExist(err) {
        if found, ok := a.findFileByName(relPath); ok { full = found }
    }
    b, err := os.ReadFile(full)
    if err != nil { return "", fmt.Errorf("nie znaleziono: %s (szukano: %s)", relPath, full) }
    ext := strings.ToLower(filepath.Ext(full))
    mime := "image/png"
    switch ext {
    case ".jpg", ".jpeg": mime = "image/jpeg"
    case ".webp": mime = "image/webp"
    }
    return fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(b)), nil
}
func (a *App) DeleteAsset(relPath string) error {
    if a.projectPath == "" { return fmt.Errorf("brak projektu") }
    relPath = a.cleanAssetPath(relPath)
    target := filepath.Join(a.projectPath, relPath)
    clean := filepath.Clean(target)
    if !a.isInsideProject(clean) { return fmt.Errorf("delete blocked") }
    return os.Remove(clean)
}
func (a *App) ListAssets(projectPath string) ([]string, error) {
    if projectPath == "" { projectPath = a.projectPath }
    if projectPath == "" { return []string{}, nil }
    var out []string
    imagesDir := filepath.Join(projectPath, "images")
    _ = filepath.WalkDir(imagesDir, func(p string, d fs.DirEntry, err error) error {
        if err != nil || d.IsDir() { return nil }
        low := strings.ToLower(d.Name())
        if strings.HasSuffix(low, ".webp") || strings.HasSuffix(low, ".png") || strings.HasSuffix(low, ".jpg") || strings.HasSuffix(low, ".jpeg") {
            rel, _ := filepath.Rel(projectPath, p)
            out = append(out, filepath.ToSlash(rel))
        }
        return nil
    })
    return out, nil
}
