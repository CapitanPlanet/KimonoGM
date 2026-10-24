package main

import (
        "fmt"
        "os"
        "path/filepath"
        "strings"
)

func (a *App) ReadJSON(fullPath string) (string, error) {
        safePath, err := a.isAllowedRoot(fullPath)
        if err != nil {
                return "", err
        }
        b, err := os.ReadFile(safePath)
        if err != nil {
                return "", fmt.Errorf("ReadJSON %s: %w", filepath.Base(safePath), err)
        }
        return string(b), nil
}

func (a *App) WriteJSON(fullPath string, content string) error {
        safePath, err := a.isAllowedRoot(fullPath)
        if err != nil {
                return err
        }
        if err := os.MkdirAll(filepath.Dir(safePath), 0755); err != nil {
                return err
        }
        return os.WriteFile(safePath, []byte(content), 0644)
}

func (a *App) SaveJsonFile(filename string, content string) error {
        if a.projectPath == "" {
                return fmt.Errorf("brak projectPath")
        }
        base := filepath.Base(filename)
        if strings.Contains(base, "..") || filepath.IsAbs(base) {
                return fmt.Errorf("invalid filename: %s", filename)
        }
        target, err := a.safeJoin(a.projectPath, "Data", base)
        if err != nil {
                return err
        }
        if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
                return err
        }
        return os.WriteFile(target, []byte(content), 0644)
}

func (a *App) ListFiles(dirPath string, ext string) ([]string, error) {
        safePath, err := a.isAllowedRoot(dirPath)
        if err != nil {
                return nil, err
        }
        entries, err := os.ReadDir(safePath)
        if err != nil {
                return nil, fmt.Errorf("ListFiles %s: %w", filepath.Base(safePath), err)
        }
        var out []string
        for _, e := range entries {
                if !e.IsDir() && (ext == "" || strings.HasSuffix(strings.ToLower(e.Name()), strings.ToLower(ext))) {
                        out = append(out, e.Name())
                }
        }
        return out, nil
}

func (a *App) DeleteFile(projectPath string, relativePath string) error {
        if projectPath == "" {
                projectPath = a.projectPath
        }
        if projectPath == "" {
                return fmt.Errorf("brak projektu")
        }
        cleaned := a.cleanAssetPath(relativePath)
        target, err := a.safeJoin(projectPath, cleaned)
        if err != nil {
                return fmt.Errorf("delete blocked: %w", err)
        }
        return os.Remove(target)
}

// FIXED: lista wszystkich dozwolonych rootów - naprawia błąd z AppData/Roaming/KimonoGM
func (a *App) getAllowedRoots() []string {
        roots := []string{}
        if a.projectPath != "" {
                roots = append(roots, a.projectPath)
        }
        if a.gamesRoot != "" {
                roots = append(roots, a.gamesRoot)
        }
        if configDir, err := os.UserConfigDir(); err == nil {
                roots = append(roots, filepath.Join(configDir, "KimonoGM"))
        }
        if home, err := os.UserHomeDir(); err == nil {
                roots = append(roots, filepath.Join(home, "Documents", "KimonoGM"))
                roots = append(roots, filepath.Join(home, "AppData", "Roaming", "KimonoGM"))
                roots = append(roots, home)
        }
        roots = append(roots, os.TempDir())
        seen := map[string]bool{}
        var uniq []string
        for _, r := range roots {
                abs, _ := filepath.Abs(r)
                if !seen[abs] && abs != "" {
                        seen[abs] = true
                        uniq = append(uniq, abs)
                }
        }
        return uniq
}

func (a *App) isAllowedRoot(fullPath string) (string, error) {
        if fullPath == "" {
                return "", fmt.Errorf("empty path")
        }
        absPath, err := filepath.Abs(filepath.Clean(fullPath))
        if err != nil {
                return "", err
        }
        for _, root := range a.getAllowedRoots() {
                absRoot, err := filepath.Abs(root)
                if err != nil {
                        continue
                }
                if absPath == absRoot {
                        return absPath, nil
                }
                rel, err := filepath.Rel(absRoot, absPath)
                if err != nil {
                        continue
                }
                if rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && !filepath.IsAbs(rel) {
                        return absPath, nil
                }
        }
        return "", fmt.Errorf("access denied: %s outside allowed roots %v", fullPath, a.getAllowedRoots())
}
