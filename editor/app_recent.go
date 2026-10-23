package main

import (
        "encoding/json"
        "os"
        "path/filepath"
)



func (a *App) recentFilePath() string {
        if configDir, err := os.UserConfigDir(); err == nil {
                return filepath.Join(configDir, "KimonoGM", "recent.json")
        }
        return filepath.Join(os.TempDir(), "kimonogm_recent.json")
}


func (a *App) GetRecentProjects() []RecentProject {
        path := a.recentFilePath()
        b, err := os.ReadFile(path)
        if err != nil {
                return []RecentProject{}
        }
        var list []RecentProject
        if err := json.Unmarshal(b, &list); err != nil {
                return []RecentProject{}
        }
        var valid []RecentProject
        for _, r := range list {
                if _, err := os.Stat(r.Path); err == nil {
                        valid = append(valid, r)
                }
        }
        return valid
}


func (a *App) AddRecentProject(projectPath string) error {
        projectPath = filepath.Clean(projectPath)
        name := filepath.Base(projectPath)
        list := a.GetRecentProjects()
        newList := []RecentProject{{Path: projectPath, Name: name}}
        for _, r := range list {
                if r.Path != projectPath && len(newList) < 10 {
                        newList = append(newList, r)
                }
        }
        b, _ := json.MarshalIndent(newList, "", "  ")
        _ = os.MkdirAll(filepath.Dir(a.recentFilePath()), 0755)
        return os.WriteFile(a.recentFilePath(), b, 0644)
}


func (a *App) ClearRecentProjects() error { return os.Remove(a.recentFilePath()) }

// ---------- SAVE META ----------


func (a *App) SaveProjectMeta(projectPath string, metaJson string) error {
        if projectPath == "" {
                projectPath = a.projectPath
        }
        var data interface{}
        if err := json.Unmarshal([]byte(metaJson), &data); err != nil {
                return err
        }
        outPath := filepath.Join(projectPath, "project.janproj")
        b, _ := json.MarshalIndent(data, "", "  ")
        return os.WriteFile(outPath, b, 0644)
}

// ---------- CREATE PROJECT ----------

