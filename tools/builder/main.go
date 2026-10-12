package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type FileEntry struct {
	RelPath string `json:"relPath"`
	FullPath string `json:"fullPath"`
	Ext string `json:"ext"`
}

type Manifest struct {
	Scenes []map[string]interface{} `json:"scenes"`
	Files []FileEntry `json:"files"`
	Avatars []map[string]interface{} `json:"avatars"`
	Sounds []map[string]interface{} `json:"sounds"`
	Stats []map[string]interface{} `json:"stats"`
	Project map[string]interface{} `json:"project"`
	Variables []map[string]interface{} `json:"variables"`
}

func main() {
	src := flag.String("src", "", "source game dir")
	out := flag.String("out", "", "out content.db")
	mode := flag.String("mode", "example", "mode")
	flag.Parse()
	if *src == "" || *out == "" {
		log.Fatal("uzyj: go run. --src=../../games/10005 --out=../../engine/BlazorWasm/wwwroot/content.db")
	}
	absSrc, _ := filepath.Abs(*src)
	fmt.Printf("[builder] mode=%s src=%s (abs %s) out=%s\n", *mode, *src, absSrc, *out)

	m := Manifest{
		Scenes: []map[string]interface{}{},
		Files: []FileEntry{},
		Avatars: []map[string]interface{}{},
		Sounds: []map[string]interface{}{},
		Stats: []map[string]interface{}{},
		Variables: []map[string]interface{}{},
	}

	// SCENES
	dataDir := filepath.Join(absSrc, "Data")
	filepath.WalkDir(dataDir, func(p string, d fs.DirEntry, err error) error {
		if err!= nil || d.IsDir() ||!strings.HasSuffix(p, ".json") {
			return nil
		}
		if strings.Contains(p, "_manifest") {
			return nil
		}
		b, _ := os.ReadFile(p)
		var arr []map[string]interface{}
		var obj map[string]interface{}
		if err := json.Unmarshal(b, &arr); err == nil {
			m.Scenes = append(m.Scenes, arr...)
		} else if err := json.Unmarshal(b, &obj); err == nil {
			if scenes, ok := obj["scenes"].([]interface{}); ok {
				for _, s := range scenes {
					if mm, ok := s.(map[string]interface{}); ok {
						m.Scenes = append(m.Scenes, mm)
					}
				}
			} else {
				m.Scenes = append(m.Scenes, obj)
			}
		}
		return nil
	})

	// PROJECT
	if b, err := os.ReadFile(filepath.Join(absSrc, "project.janproj")); err == nil {
		json.Unmarshal(b, &m.Project)
	}

	// FILES - FIX REL
	for _, sub := range []string{"images", "sounds"} {
		root := filepath.Join(absSrc, sub)
		filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err!= nil || d.IsDir() {
				return nil
			}
			absP, _ := filepath.Abs(p)
			rel, _ := filepath.Rel(absSrc, absP)
			rel = filepath.ToSlash(rel)

			m.Files = append(m.Files, FileEntry{RelPath: rel, FullPath: absP, Ext: filepath.Ext(p)})

			if sub == "sounds" {
				sType := "sfx"
				if strings.Contains(rel, "music") {
					sType = "music"
				} else if strings.Contains(rel, "voice") {
					sType = "voice"
				}
				m.Sounds = append(m.Sounds, map[string]interface{}{
					"_id": rel, "type": sType, "blobId": rel, "volume": 1.0,
				})
			}
			return nil
		})
	}

	// AVATARS
	avatarMap := map[string]map[string]string{}
	for _, f := range m.Files {
		if!strings.HasPrefix(f.RelPath, "images/") {
			continue
		}
		base := filepath.Base(f.RelPath)
		if strings.HasPrefix(base, "janusz") {
			parts := strings.Split(strings.TrimSuffix(base, f.Ext), "_")
			emotion := "neutral"
			if len(parts) > 1 {
				emotion = parts[len(parts)-1]
			}
			if _, ok := avatarMap["janusz"];!ok {
				avatarMap["janusz"] = map[string]string{}
			}
			avatarMap["janusz"][emotion] = f.RelPath
		}
	}
	for id, emotions := range avatarMap {
		m.Avatars = append(m.Avatars, map[string]interface{}{
			"_id": id, "name": "Janusz", "emotions": emotions,
		})
	}

	// STATS
	if b, err := os.ReadFile(filepath.Join(absSrc, "stats.json")); err == nil {
		var arr []map[string]interface{}
		json.Unmarshal(b, &arr)
		m.Stats = arr
	} else {
		m.Stats = []map[string]interface{}{
			{"_id": "money", "label": "Kasa", "default": 100, "min": 0, "max": 9999},
			{"_id": "stress", "label": "Stres", "default": 0, "min": 0, "max": 100},
			{"_id": "respect", "label": "Szacunek", "default": 50, "min": 0, "max": 100},
		}
	}

	fmt.Printf("[builder] znaleziono %d scen, %d blobow, %d avatary, %d dzwieki, %d staty\n",
		len(m.Scenes), len(m.Files), len(m.Avatars), len(m.Sounds), len(m.Stats))
	for _, f := range m.Files {
		fmt.Printf(" file: %s\n", f.RelPath)
	}

	tmpManifest := filepath.Join(os.TempDir(), "kimono_manifest.json")
	b, _ := json.MarshalIndent(m, "", " ")
	os.WriteFile(tmpManifest, b, 0644)

	packerDir := "../packer"
	if _, err := os.Stat("../../tools/packer/packer.csproj"); err == nil {
		packerDir = "../../tools/packer"
	}
	cmd := exec.Command("dotnet", "run", "--project", filepath.Join(packerDir, "packer.csproj"), "--", tmpManifest, *out)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println("[builder] wywoluje packer.NET...")
	if err := cmd.Run(); err!= nil {
		log.Fatalf("packer failed: %v", err)
	}
	fmt.Printf("[builder] OK -> %s\n", *out)
}