package bohriumskills

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmbeddedNativePackageInventory(t *testing.T) {
	data, err := os.ReadFile("cli-skills.lock.json")
	if err != nil {
		t.Fatal(err)
	}
	var lock struct {
		Skills []struct {
			Name  string   `json:"name"`
			Files []string `json:"files"`
		} `json:"skills"`
	}
	if err := json.Unmarshal(data, &lock); err != nil {
		t.Fatal(err)
	}
	for _, lang := range []string{"zh", "en"} {
		skills, err := OfficialSkills(lang)
		if err != nil {
			t.Fatal(err)
		}
		byName := make(map[string]Skill)
		for _, skill := range skills {
			byName[skill.Name] = skill
		}
		if len(skills) != len(lock.Skills) {
			t.Fatalf("%s: embedded Skill count differs from source lock", lang)
		}
		for _, item := range lock.Skills {
			skill, ok := byName[item.Name]
			if !ok || strings.Join(skill.Files, "\n") != strings.Join(item.Files, "\n") {
				t.Errorf("%s/%s: embedded files = %v, source inventory = %v", lang, item.Name, skill.Files, item.Files)
			}
			delete(byName, item.Name)
		}
		if len(byName) != 0 {
			t.Errorf("%s: unexpected embedded Skills: %v", lang, byName)
		}
	}
}

func TestMarketplaceSkillCopiesMatchEnglish(t *testing.T) {
	skills, err := OfficialSkills("en")
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) == 0 {
		t.Fatal("no English Skills found")
	}
	for _, skill := range skills {
		t.Run(skill.Name, func(t *testing.T) {
			root := filepath.Join("skills", skill.Name)
			want := make(map[string]bool, len(skill.Files))
			for _, name := range skill.Files {
				want[name] = true
			}
			err := filepath.WalkDir(root, func(file string, entry fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if entry.Type()&os.ModeSymlink != 0 {
					t.Errorf("%s must not be a symbolic link", file)
					return nil
				}
				if entry.IsDir() {
					return nil
				}
				if !entry.Type().IsRegular() {
					t.Errorf("%s is not a regular file", file)
					return nil
				}
				rel, err := filepath.Rel(root, file)
				if err != nil {
					return err
				}
				rel = filepath.ToSlash(rel)
				if !want[rel] {
					t.Errorf("unexpected file %s", file)
					return nil
				}
				delete(want, rel)
				got, err := os.ReadFile(file)
				if err != nil {
					return err
				}
				expected, err := EmbeddedFS.ReadFile(path.Join(skill.Path, rel))
				if err != nil {
					return err
				}
				if !bytes.Equal(got, expected) {
					t.Errorf("%s differs from %s", file, path.Join(skill.Path, rel))
				}
				return nil
			})
			if err != nil {
				t.Error(err)
			}
			for name := range want {
				t.Errorf("%s missing %s", root, name)
			}
		})
	}
}

func TestClaudeMarketplaceResolvesPublishedSkills(t *testing.T) {
	data, err := os.ReadFile(".claude-plugin/marketplace.json")
	if err != nil {
		t.Fatal(err)
	}
	var marketplace struct {
		Name    string `json:"name"`
		Plugins []struct {
			Name   string   `json:"name"`
			Source string   `json:"source"`
			Skills []string `json:"skills"`
		} `json:"plugins"`
	}
	if err := json.Unmarshal(data, &marketplace); err != nil {
		t.Fatal(err)
	}
	if marketplace.Name != "bohrium" || len(marketplace.Plugins) != 1 {
		t.Fatalf("marketplace identity changed: %+v", marketplace)
	}
	plugin := marketplace.Plugins[0]
	if plugin.Name != "bohrium-skills" || plugin.Source != "./" {
		t.Fatalf("plugin identity/source changed: %+v", plugin)
	}
	skills, err := OfficialSkills("en")
	if err != nil {
		t.Fatal(err)
	}
	want := make(map[string]bool)
	for _, skill := range skills {
		want["./"+skill.Path] = true
	}
	setup, err := os.ReadFile("skills/bohrium-setup/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	setupName, err := frontmatterName(string(setup))
	if err != nil || setupName != "bohrium-setup" {
		t.Fatalf("invalid setup Skill name %q: %v", setupName, err)
	}
	want["./skills/bohrium-setup"] = true
	for _, path := range plugin.Skills {
		if !want[path] {
			t.Errorf("unknown or duplicate marketplace Skill %q", path)
		}
		delete(want, path)
	}
	if len(want) != 0 {
		t.Errorf("Skills missing from marketplace: %v", want)
	}
}
