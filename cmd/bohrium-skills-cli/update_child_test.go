package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestUpdatedBinaryIsResolvedAndVerifiedBeforeSync(t *testing.T) {
	fixtureRoot := t.TempDir()
	code := `package main
import("encoding/json";"fmt";"os";"path/filepath";"strings")
func main(){
 exe,_:=os.Executable()
 if strings.TrimSuffix(filepath.Base(exe),".exe")=="npm" { fmt.Println(os.Getenv("FIXTURE_NPM_ROOT")); return }
 fresh:=exe==os.Getenv("FIXTURE_FRESH_BINARY")
 version:="0.1.0"; if fresh { version=os.Getenv("FIXTURE_NEW_VERSION") }
 if len(os.Args)>1 && os.Args[1]=="version" { json.NewEncoder(os.Stdout).Encode(map[string]string{"version":version});return }
 if !fresh { os.WriteFile(os.Getenv("FIXTURE_SHADOW_MARKER"),[]byte("old binary used"),0600);return }
 lang:="zh";for i,a:=range os.Args {if a=="--lang" && i+1<len(os.Args){lang=os.Args[i+1]}}
 os.WriteFile(os.Getenv("FIXTURE_SYNC_MARKER"),[]byte(version),0600)
 receiptVersion:=os.Getenv("FIXTURE_RECEIPT_VERSION");if receiptVersion=="" {receiptVersion=version}
 json.NewEncoder(os.Stdout).Encode(map[string]string{"action":"synced","version":receiptVersion,"lang":lang})
}`
	source := filepath.Join(fixtureRoot, "fixture.go")
	if err := os.WriteFile(source, []byte(code), 0o644); err != nil {
		t.Fatal(err)
	}
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	helper := filepath.Join(fixtureRoot, "fixture"+ext)
	build := exec.Command("go", "build", "-o", helper, source)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build process fixture: %v\n%s", err, out)
	}
	binary, err := os.ReadFile(helper)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name           string
		version        string
		receiptVersion string
		wantError      bool
		wantSync       bool
	}{
		{"wrong binary version", "0.1.0", "", true, false},
		{"wrong sync receipt", "0.2.0", "0.1.0", true, true},
		{"fresh binary despite PATH shadow", "0.2.0", "", false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			base := t.TempDir()
			binDir := filepath.Join(base, "bin")
			npmRoot := filepath.Join(base, "lib", "node_modules")
			packageDir := filepath.Join(npmRoot, "bohrium-skills-cli")
			fresh := filepath.Join(packageDir, "bin", "bohrium-skills-cli"+ext)
			for _, file := range []string{filepath.Join(binDir, "npm"+ext), filepath.Join(binDir, "bohrium-skills-cli"+ext), fresh} {
				if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(file, binary, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			manifest, _ := json.Marshal(map[string]string{"name": "bohrium-skills-cli", "version": "0.2.0"})
			if err := os.WriteFile(filepath.Join(packageDir, "package.json"), manifest, 0o644); err != nil {
				t.Fatal(err)
			}
			shadowMarker, syncMarker := filepath.Join(base, "shadow-used"), filepath.Join(base, "fresh-synced")
			t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("HOME", base)
			t.Setenv("FIXTURE_NPM_ROOT", npmRoot)
			t.Setenv("FIXTURE_FRESH_BINARY", fresh)
			t.Setenv("FIXTURE_NEW_VERSION", tc.version)
			t.Setenv("FIXTURE_RECEIPT_VERSION", tc.receiptVersion)
			t.Setenv("FIXTURE_SHADOW_MARKER", shadowMarker)
			t.Setenv("FIXTURE_SYNC_MARKER", syncMarker)
			// An old internal guard must not bypass child-version verification.
			t.Setenv("BOHRIUM_SKILLS_CLI_CHILD_SYNC", "1")
			withVersion(t, "0.1.0")
			err := syncForVersion("0.2.0", "en", false)
			if (err != nil) != tc.wantError {
				t.Fatalf("sync error = %v, want error %v", err, tc.wantError)
			}
			if _, err := os.Stat(shadowMarker); !os.IsNotExist(err) {
				t.Fatalf("PATH-shadowed old binary was invoked: %v", err)
			}
			_, err = os.Stat(syncMarker)
			if (err == nil) != tc.wantSync {
				t.Fatalf("new child sync presence = %v, want %v", err == nil, tc.wantSync)
			}
		})
	}
}
