package syncer

import (
	"errors"
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
