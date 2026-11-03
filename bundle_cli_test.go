package bohriumskills

import (
	"encoding/json"
	"os"
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
