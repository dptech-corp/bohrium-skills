package bohriumskills

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

var guideReadPattern = regexp.MustCompile(`(?m)^bohr guide read ([a-z0-9-]+)(?: ([a-zA-Z0-9_./-]+))? -o text$`)
var helpPattern = regexp.MustCompile(`(?m)^bohr ([a-z][a-z -]*?) --help$`)

func TestPublishedSkillsUseCLI(t *testing.T) {
	guides := make(map[string]string)
	for _, lang := range []string{"zh", "en"} {
		skills, err := OfficialSkills(lang)
		if err != nil {
			t.Fatal(err)
		}
		for _, skill := range skills {
			t.Run(lang+"/"+skill.Name, func(t *testing.T) {
				if len(skill.Files) != 1 || skill.Files[0] != "SKILL.md" {
					t.Fatalf("published Skill still bundles API resources: %v", skill.Files)
				}
				data, err := fs.ReadFile(EmbeddedFS, skill.Path+"/SKILL.md")
				if err != nil {
					t.Fatal(err)
				}
				text := string(data)
				for _, legacy := range []string{"/openapi/", "open.bohrium.com", "openapi.dp.tech", "requests.", "curl ", "export ACCESS_KEY", "bohr skills ", "~/.openclaw/"} {
					if strings.Contains(text, legacy) {
						t.Errorf("contains legacy API or credential instruction %q", legacy)
					}
				}
				for _, required := range []string{"bohr --version", "npm install -g @dptech-corp/bohr-cli@latest", "bohr auth status -o json", "bohr auth login"} {
					if !strings.Contains(text, required) {
						t.Errorf("missing CLI setup step %q", required)
					}
				}
				reads := guideReadPattern.FindAllStringSubmatch(text, -1)
				if len(reads) < 2 || reads[0][2] != "" {
					t.Fatal("must load an embedded guide and a task reference")
				}
				guide := reads[0][1]
				if previous, ok := guides[skill.Name]; ok && previous != guide {
					t.Errorf("language mappings differ: %q and %q", previous, guide)
				}
				guides[skill.Name] = guide
				if !strings.Contains(text, "bohr guide list "+guide+"/references -o json") {
					t.Error("missing embedded reference discovery command")
				}
				if len(helpPattern.FindAllStringSubmatch(text, -1)) == 0 {
					t.Error("missing domain command help")
				}
			})
		}
	}
	if len(guides) != 17 {
		t.Fatalf("language trees expose different Skill names: %v", guides)
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

func TestPublishedGuidanceAgainstBohr(t *testing.T) {
	bin := os.Getenv("BOHR_SKILLS_TEST_BIN")
	if bin == "" {
		t.Skip("set BOHR_SKILLS_TEST_BIN to a published bohr binary; required by CI")
	}
	t.Setenv("BOHR_CONFIG_DIR", t.TempDir())
	t.Setenv("BOHR_ACCESS_KEY", "")
	run := func(t *testing.T, args ...string) []byte {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, bin, args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("bohr %v: %v\n%s\n%s", args, err, out, stderr.String())
		}
		return out
	}
	seen := make(map[string]bool)
	checkedHelp := make(map[string]bool)
	for _, lang := range []string{"zh", "en"} {
		skills, err := OfficialSkills(lang)
		if err != nil {
			t.Fatal(err)
		}
		for _, skill := range skills {
			data, err := fs.ReadFile(EmbeddedFS, skill.Path+"/SKILL.md")
			if err != nil {
				t.Fatal(err)
			}
			for _, match := range guideReadPattern.FindAllStringSubmatch(string(data), -1) {
				guide, reference := match[1], match[2]
				key := guide + "/" + reference
				if seen[key] {
					continue
				}
				seen[key] = true
				t.Run(key, func(t *testing.T) {
					args := []string{"guide", "read", guide}
					if reference != "" {
						args = append(args, reference)
					}
					out := run(t, append(args, "-o", "json")...)
					var result struct {
						OK   bool `json:"ok"`
						Data struct {
							Name    string `json:"name"`
							Path    string `json:"path"`
							Content string `json:"content"`
						} `json:"data"`
					}
					if err := json.Unmarshal(out, &result); err != nil {
						t.Fatalf("invalid guide JSON: %v\n%s", err, out)
					}
					path := reference
					if path == "" {
						path = "SKILL.md"
					}
					if !result.OK || result.Data.Name != guide || result.Data.Path != path || strings.TrimSpace(result.Data.Content) == "" {
						t.Fatalf("incorrect guide identity or empty content: %s", out)
					}
					wantText := result.Data.Content
					if !strings.HasSuffix(wantText, "\n") {
						wantText += "\n"
					}
					if text := string(run(t, append(args, "-o", "text")...)); text != wantText {
						t.Fatal("text guide read differs from JSON content")
					}
					if reference == "" {
						out := run(t, "guide", "list", guide+"/references", "-o", "json")
						var listing struct {
							OK   bool `json:"ok"`
							Data struct {
								Path  string            `json:"path"`
								Items []json.RawMessage `json:"items"`
							} `json:"data"`
						}
						if err := json.Unmarshal(out, &listing); err != nil {
							t.Fatalf("invalid reference listing: %v\n%s", err, out)
						}
						if !listing.OK || listing.Data.Path != guide+"/references" || len(listing.Data.Items) == 0 {
							t.Fatalf("incorrect or empty reference listing: %s", out)
						}
					}
				})
			}
			for _, match := range helpPattern.FindAllStringSubmatch(string(data), -1) {
				if !checkedHelp[match[1]] {
					run(t, append(strings.Fields(match[1]), "--help")...)
					checkedHelp[match[1]] = true
				}
			}
		}
	}
	if len(seen) < 34 {
		t.Fatalf("checked %d guide/reference targets, want at least 34", len(seen))
	}
	project := t.TempDir()
	run(t, "skill", "install", "./en/bohrium-paper-search", "--target", "codex", "--project-dir", project, "-o", "json")
	installed, err := os.ReadFile(filepath.Join(project, ".agents", "skills", "bohrium-paper-search", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := fs.ReadFile(EmbeddedFS, "en/bohrium-paper-search/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(installed) != string(source) {
		t.Fatal("published bohr installer changed Skill content")
	}
}
