package updater

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInstallMethodUsesPackageProvenance(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	cases := []struct {
		name        string
		dir         string
		packageName string
		want        InstallMethod
	}{
		{"documented manual binary", "local/bin", "", InstallManual},
		{"misleading module directory", "not-node_modules/bohrium-skills-cli/bin", NpmPackage, InstallManual},
		{"missing package identity", "node_modules/bohrium-skills-cli/bin", "", InstallManual},
		{"wrong package identity", "node_modules/bohrium-skills-cli/bin", "different-package", InstallManual},
		{"npm global package", "lib/node_modules/bohrium-skills-cli/bin", NpmPackage, InstallNPM},
		{"pnpm package", "node_modules/.pnpm/bohrium-skills-cli@1.2.0/node_modules/bohrium-skills-cli/bin", NpmPackage, InstallNPM},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), filepath.FromSlash(tc.dir))
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(dir, NpmPackage+ext)
			if err := os.WriteFile(binary, data, 0o755); err != nil {
				t.Fatal(err)
			}
			if tc.packageName != "" {
				manifest, _ := json.Marshal(map[string]string{"name": tc.packageName, "version": "1.2.0"})
				if err := os.WriteFile(filepath.Join(filepath.Dir(dir), "package.json"), manifest, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command(binary, "-test.run=^TestInstallMethodHelper$")
			cmd.Env = append(os.Environ(), "BOHRIUM_INSTALL_METHOD_HELPER=1")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("classification helper: %v\n%s", err, out)
			}
			if got := strings.TrimSpace(string(out)); got != string(tc.want) {
				t.Fatalf("install method = %q, want %q", got, tc.want)
			}
			if tc.name == "npm global package" {
				link := filepath.Join(t.TempDir(), NpmPackage+ext)
				if err := os.Symlink(binary, link); err != nil {
					t.Skipf("host does not permit launcher symlinks: %v", err)
				}
				linked := exec.Command(link, "-test.run=^TestInstallMethodHelper$")
				linked.Env = cmd.Env
				out, err := linked.CombinedOutput()
				if err != nil || strings.TrimSpace(string(out)) != string(InstallNPM) {
					t.Fatalf("symlinked npm launcher misclassified: %v %s", err, out)
				}
			}
		})
	}
}

func TestInstallMethodHelper(t *testing.T) {
	if os.Getenv("BOHRIUM_INSTALL_METHOD_HELPER") != "1" {
		return
	}
	fmt.Print(RealRunner{}.DetectInstallMethod())
	os.Exit(0)
}
