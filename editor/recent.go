package main

import (
    "encoding/json"
    "os"
    "path/filepath"
)

func (a *App) recentFilePath() string {
    home, _ := os.UserHomeDir()
    dir := filepath.Join(home, ".janusz-maker")
    _ = os.MkdirAll(dir, 0755)
    return filepath.Join(dir, "recent.json")
}
func (a *App) GetRecentProjects() ([]string, error) {
    path := a.recentFilePath()
    b, err := os.ReadFile(path)
    if err != nil { return []string{}, nil }
    var list []string
    if err := json.Unmarshal(b, &list); err != nil { return []string{}, nil }
    var out []string
    for _, p := range list {
        if _, err := os.Stat(p); err == nil { out = append(out, p) }
        if len(out) >= 10 { break }
    }
    return out, nil
}
func (a *App) AddRecentProject(p string) error {
    p = filepath.Clean(p)
    if p == "" { return nil }
    existing, _ := a.GetRecentProjects()
    var filtered []string
    for _, e := range existing { if e != p { filtered = append(filtered, e) } }
    filtered = append([]string{p}, filtered...)
    if len(filtered) > 10 { filtered = filtered[:10] }
    b, _ := json.Marshal(filtered)
    return os.WriteFile(a.recentFilePath(), b, 0644)
}
