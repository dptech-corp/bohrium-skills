package syncer

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

func nativeFixture(label string) fs.FS {
	return fstest.MapFS{
		"en/bohrium-job/SKILL.md":               {Data: []byte("---\nname: bohrium-job\ndescription: Job\n---\n" + label + "\n")},
		"en/bohrium-job/references/commands.md": {Data: []byte(label + " reference\n")},
		"en/bohrium-job/agents/openai.yaml":     {Data: []byte(label + " interface\n")},
	}
}

func writeTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, text := range files {
		file := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func readTree(t *testing.T, root string) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = string(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestSameClockSyncKeepsOriginalAndIntermediateBackups(t *testing.T) {
	home := t.TempDir()
	config := filepath.Join(home, "config")
	target := filepath.Join(home, ".agents", "skills")
	dest := filepath.Join(target, "bohrium-job")
	writeTree(t, dest, map[string]string{"SKILL.md": "original API snapshot\n", "personal-note.txt": "user edit\n"})
	fixed := func() time.Time { return time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC) }
	for _, label := range []string{"first native", "second native"} {
		_, err := Sync(SyncOptions{Version: label, Lang: "en", HomeDir: home, ConfigDir: config, Targets: []string{target}, SourceFS: nativeFixture(label), Force: true, Now: fixed})
		if err != nil {
			t.Fatal(err)
		}
	}
	var snapshots []string
	var notes []string
	err := filepath.WalkDir(filepath.Join(config, "backups"), func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Name() == "SKILL.md" || entry.Name() == "personal-note.txt" {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if entry.Name() == "SKILL.md" {
				snapshots = append(snapshots, string(data))
			} else {
				notes = append(notes, string(data))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	wantFirst := "---\nname: bohrium-job\ndescription: Job\n---\nfirst native\n"
	foundOriginal, foundFirst := false, false
	for _, snapshot := range snapshots {
		foundOriginal = foundOriginal || snapshot == "original API snapshot\n"
		foundFirst = foundFirst || snapshot == wantFirst
	}
	if len(snapshots) != 2 || !foundOriginal || !foundFirst || !reflect.DeepEqual(notes, []string{"user edit\n"}) {
		t.Fatalf("backups were overwritten: snapshots=%q notes=%q", snapshots, notes)
	}
}

func TestReplacementFailureNeverLeavesPartialActiveTree(t *testing.T) {
	home := t.TempDir()
	defer filepath.WalkDir(home, func(path string, entry os.DirEntry, err error) error {
		if err == nil && entry.IsDir() {
			_ = os.Chmod(path, 0o755)
		}
		return nil
	})
	target := filepath.Join(home, "skills")
	dest := filepath.Join(target, "bohrium-job")
	old := map[string]string{
		"SKILL.md": "old main\n", "references/commands.md": "old reference\n",
		"agents/openai.yaml": "old interface\n", "user-note.txt": "user data\n",
	}
	writeTree(t, dest, old)
	if err := os.Chmod(filepath.Join(dest, "references"), 0o500); err != nil {
		t.Fatal(err)
	}
	_, syncErr := Sync(SyncOptions{Version: "new", Lang: "en", HomeDir: home, ConfigDir: filepath.Join(home, "config"), Targets: []string{target}, SourceFS: nativeFixture("new"), Force: true})
	actual := readTree(t, dest)
	newTree := map[string]string{
		"SKILL.md":               "---\nname: bohrium-job\ndescription: Job\n---\nnew\n",
		"references/commands.md": "new reference\n", "agents/openai.yaml": "new interface\n",
	}
	if !reflect.DeepEqual(actual, old) && !reflect.DeepEqual(actual, newTree) {
		t.Fatalf("sync left a partial active tree after %v: %v", syncErr, actual)
	}
}

func TestPublishFailureRestoresCompletePreviousTree(t *testing.T) {
	root := t.TempDir()
	active, staged := filepath.Join(root, "active"), filepath.Join(root, "staged")
	old := map[string]string{"SKILL.md": "old", "references/commands.md": "old ref", "agents/openai.yaml": "old interface"}
	writeTree(t, active, old)
	writeTree(t, staged, map[string]string{"SKILL.md": "new", "references/commands.md": "new ref"})
	publishErr := errors.New("injected publication failure")
	err := publishSkillDirectory(staged, active, func(from, to string) error {
		if from == staged && to == active {
			return publishErr
		}
		return os.Rename(from, to)
	})
	if !errors.Is(err, publishErr) {
		t.Fatalf("publication error was lost: %v", err)
	}
	if got := readTree(t, active); !reflect.DeepEqual(got, old) {
		t.Fatalf("old installation was not restored intact: %v", got)
	}
}

func TestRollbackFailureRetainsRecoveryTreeAndBothErrors(t *testing.T) {
	root := t.TempDir()
	active, staged := filepath.Join(root, "active"), filepath.Join(root, "staged")
	old := map[string]string{"SKILL.md": "old", "references/commands.md": "old ref"}
	writeTree(t, active, old)
	writeTree(t, staged, map[string]string{"SKILL.md": "new"})
	publishErr, rollbackErr := errors.New("injected publication failure"), errors.New("injected rollback failure")
	previous := ""
	err := publishSkillDirectory(staged, active, func(from, to string) error {
		if from == staged && to == active {
			return publishErr
		}
		if from == previous && to == active {
			return rollbackErr
		}
		if from == active {
			previous = to
		}
		return os.Rename(from, to)
	})
	if !errors.Is(err, publishErr) || !errors.Is(err, rollbackErr) {
		t.Fatalf("publication or rollback error was lost: %v", err)
	}
	if previous == "" || !strings.Contains(err.Error(), previous) {
		t.Fatalf("recovery location was not reported: %v", err)
	}
	if got := readTree(t, previous); !reflect.DeepEqual(got, old) {
		t.Fatalf("recoverable previous tree was deleted or changed: %v", got)
	}
}

func TestDefaultInstallMigratesLegacyCodexWithoutDuplicatingSkills(t *testing.T) {
	home := t.TempDir()
	legacy := filepath.Join(home, ".codex", "skills", "bohrium-job")
	writeTree(t, legacy, map[string]string{"SKILL.md": "old API", "custom-note.txt": "keep user edit"})
	custom := filepath.Join(home, ".codex", "skills", "my-custom-skill")
	writeTree(t, custom, map[string]string{"SKILL.md": "unrelated user skill"})
	config := filepath.Join(home, "config")
	if err := WriteState(config, State{Version: "0.1.1", OfficialSkills: []string{"bohrium-job"}, TargetDirs: []string{filepath.Dir(legacy)}}); err != nil {
		t.Fatal(err)
	}
	result, err := Sync(SyncOptions{Version: "native", Lang: "en", HomeDir: home, ConfigDir: config, SourceFS: nativeFixture("new")})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.TargetDirs) != 2 || !reflect.DeepEqual(result.MigratedLegacy, []string{"bohrium-job"}) {
		t.Fatalf("incorrect migration receipt: %+v", result)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatalf("duplicate legacy copy remains: %v", err)
	}
	if got := readTree(t, custom); !reflect.DeepEqual(got, map[string]string{"SKILL.md": "unrelated user skill"}) {
		t.Fatalf("unrelated custom Skill changed: %v", got)
	}
	for _, target := range result.TargetDirs {
		got := readTree(t, filepath.Join(target, "bohrium-job"))
		if got["references/commands.md"] != "new reference\n" || got["agents/openai.yaml"] != "new interface\n" {
			t.Fatalf("canonical package incomplete: %v", got)
		}
	}
	var originalNote bool
	if err := filepath.WalkDir(filepath.Join(config, "backups"), func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Name() == "custom-note.txt" {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			originalNote = string(data) == "keep user edit"
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !originalNote {
		t.Fatal("legacy user additions were not backed up")
	}
}

func TestLegacyRootAliasNeverRetiresCanonicalTree(t *testing.T) {
	home := t.TempDir()
	canonical := filepath.Join(home, ".agents", "skills")
	if err := os.MkdirAll(canonical, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, ".codex"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(canonical, filepath.Join(home, ".codex", "skills")); err != nil {
		t.Skipf("host does not permit directory symlinks: %v", err)
	}
	result, err := Sync(SyncOptions{Version: "new", Lang: "en", HomeDir: home, ConfigDir: filepath.Join(home, "config"), SourceFS: nativeFixture("new")})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.MigratedLegacy) != 0 {
		t.Fatalf("canonical aliases were retired: %v", result.MigratedLegacy)
	}
	files := readTree(t, filepath.Join(canonical, "bohrium-job"))
	if files["SKILL.md"] != "---\nname: bohrium-job\ndescription: Job\n---\nnew\n" || len(files) != 3 {
		t.Fatalf("canonical tree was lost through legacy alias: %v", files)
	}
}

func TestLegacyRootAliasesToOtherTreesArePreserved(t *testing.T) {
	for _, targetName := range []string{".claude/skills", "shared-source"} {
		t.Run(targetName, func(t *testing.T) {
			home := t.TempDir()
			target := filepath.Join(home, filepath.FromSlash(targetName))
			old := map[string]string{"bohrium-job/SKILL.md": "original user content", "my-custom/SKILL.md": "custom"}
			writeTree(t, target, old)
			if err := os.MkdirAll(filepath.Join(home, ".codex"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, filepath.Join(home, ".codex", "skills")); err != nil {
				t.Skipf("host cannot create symlinks: %v", err)
			}
			_, err := Sync(SyncOptions{Version: "new", Lang: "en", HomeDir: home, ConfigDir: filepath.Join(home, "config"), SourceFS: nativeFixture("new"), Force: true})
			if err == nil || !strings.Contains(err.Error(), "alias") {
				t.Fatalf("unsafe root alias was not reported: %v", err)
			}
			if got := readTree(t, target); !reflect.DeepEqual(got, old) {
				t.Fatalf("user shared tree changed: %v", got)
			}
		})
	}
}

func TestNonDirectoryLegacyItemsAreNeverDeletedOrClaimedAsBackups(t *testing.T) {
	for _, linked := range []bool{false, true} {
		t.Run(fmt.Sprintf("symlink=%v", linked), func(t *testing.T) {
			home := t.TempDir()
			legacy := filepath.Join(home, ".codex", "skills", "bohrium-job")
			if err := os.MkdirAll(filepath.Dir(legacy), 0o755); err != nil {
				t.Fatal(err)
			}
			if linked {
				file := filepath.Join(home, "user-file")
				if err := os.WriteFile(file, []byte("keep user data"), 0o644); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(file, legacy); err != nil {
					t.Skipf("host cannot create symlinks: %v", err)
				}
			} else if err := os.WriteFile(legacy, []byte("keep user data"), 0o644); err != nil {
				t.Fatal(err)
			}
			result, err := Sync(SyncOptions{Version: "new", Lang: "en", HomeDir: home, ConfigDir: filepath.Join(home, "config"), SourceFS: nativeFixture("new"), Force: true})
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(legacy)
			if err != nil || string(data) != "keep user data" {
				t.Fatalf("legacy user file changed: %v %q", err, data)
			}
			if len(result.MigratedLegacy) != 0 || result.BackupCount != 0 {
				t.Fatalf("non-directory claimed as migrated/backed up: %+v", result)
			}
		})
	}
}

func TestBackupRejectsNonDirectoryBeforeChangingExistingBackup(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "file")
	dest := filepath.Join(root, "backup")
	if err := os.WriteFile(source, []byte("source file"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeTree(t, dest, map[string]string{"saved": "previous backup"})
	if err := copyDir(source, dest); err == nil {
		t.Fatal("file was silently accepted as a directory backup")
	}
	if got := readTree(t, dest); !reflect.DeepEqual(got, map[string]string{"saved": "previous backup"}) {
		t.Fatalf("existing backup was changed: %v", got)
	}
}

func TestLegacyLeafAliasIsRemovedWithoutDeletingClaudeTarget(t *testing.T) {
	home := t.TempDir()
	claude := filepath.Join(home, ".claude", "skills", "bohrium-job")
	writeTree(t, claude, map[string]string{"SKILL.md": "old Claude copy", "user-note.txt": "keep"})
	legacy := filepath.Join(home, ".codex", "skills", "bohrium-job")
	if err := os.MkdirAll(filepath.Dir(legacy), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(claude, legacy); err != nil {
		t.Skipf("host cannot create symlinks: %v", err)
	}
	result, err := Sync(SyncOptions{Version: "new", Lang: "en", HomeDir: home, ConfigDir: filepath.Join(home, "config"), SourceFS: nativeFixture("new"), Force: true})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.MigratedLegacy, []string{"bohrium-job"}) {
		t.Fatalf("leaf alias was not retired: %+v", result)
	}
	if _, err := os.Lstat(legacy); !os.IsNotExist(err) {
		t.Fatalf("legacy duplicate alias remains: %v", err)
	}
	for _, target := range result.TargetDirs {
		got := readTree(t, filepath.Join(target, "bohrium-job"))
		if len(got) != 3 || got["references/commands.md"] != "new reference\n" {
			t.Fatalf("active canonical target was deleted or incomplete: %v", got)
		}
	}
}
