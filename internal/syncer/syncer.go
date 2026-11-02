package syncer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const StateFileName = "skills-state.json"

type State struct {
	Version              string   `json:"version"`
	Lang                 string   `json:"lang"`
	OfficialSkills       []string `json:"official_skills"`
	UpdatedSkills        []string `json:"updated_skills"`
	AddedOfficialSkills  []string `json:"added_official_skills"`
	SkippedDeletedSkills []string `json:"skipped_deleted_skills"`
	TargetDirs           []string `json:"target_dirs"`
	UpdatedAt            string   `json:"updated_at"`
}

type PlanInput struct {
	Version        string
	OfficialSkills []string
	LocalSkills    []string
	PreviousState  *State
	StateReadable  bool
	Force          bool
}

type Plan struct {
	Version        string
	OfficialSkills []string
	ToUpdate       []string
	Added          []string
	SkippedDeleted []string
}

type SyncOptions struct {
	Version   string
	Lang      string
	HomeDir   string
	ConfigDir string
	Targets   []string
	SourceFS  fs.FS
	Force     bool
	Now       func() time.Time
}

type Result struct {
	Action         string   `json:"action"`
	Version        string   `json:"version"`
	Lang           string   `json:"lang"`
	Official       []string `json:"official"`
	Updated        []string `json:"updated"`
	Added          []string `json:"added"`
	SkippedDeleted []string `json:"skipped_deleted"`
	BackupCount    int      `json:"backup_count"`
	TargetDirs     []string `json:"target_dirs"`
	StatePath      string   `json:"state_path"`
	MigratedLegacy []string `json:"migrated_legacy,omitempty"`
}

type sourceSkill struct {
	Name string
	Path string
}

func PlanSync(input PlanInput) Plan {
	official := uniqueSorted(input.OfficialSkills)
	if input.Force {
		return Plan{Version: input.Version, OfficialSkills: official, ToUpdate: official}
	}

	officialSet := toSet(official)
	localOfficial := intersection(input.LocalSkills, officialSet)

	var previousOfficial []string
	if input.StateReadable && input.PreviousState != nil {
		previousOfficial = input.PreviousState.OfficialSkills
	}
	previousSet := toSet(previousOfficial)

	var added []string
	for _, skill := range official {
		if !previousSet[skill] {
			added = append(added, skill)
		}
	}

	updateSet := toSet(localOfficial)
	for _, skill := range added {
		updateSet[skill] = true
	}
	toUpdate := sortedKeys(updateSet)
	updateSet = toSet(toUpdate)

	var skipped []string
	for _, skill := range official {
		if !updateSet[skill] {
			skipped = append(skipped, skill)
		}
	}

	return Plan{
		Version:        input.Version,
		OfficialSkills: official,
		ToUpdate:       toUpdate,
		Added:          added,
		SkippedDeleted: skipped,
	}
}

func Sync(opts SyncOptions) (*Result, error) {
	if opts.Lang == "" {
		opts.Lang = "zh"
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.SourceFS == nil {
		return nil, errors.New("source FS is nil")
	}
	home, err := resolveHome(opts.HomeDir)
	if err != nil {
		return nil, err
	}
	defaultTargets := len(opts.Targets) == 0
	if defaultTargets {
		opts.Targets = DefaultTargets(home)
	}
	configDir := opts.ConfigDir
	if configDir == "" {
		configDir = filepath.Join(home, ".config", "bohrium-skills-cli")
	}

	sources, err := listSourceSkills(opts.SourceFS, opts.Lang)
	if err != nil {
		return nil, err
	}
	sourceByName := map[string]sourceSkill{}
	var official []string
	for _, source := range sources {
		official = append(official, source.Name)
		sourceByName[source.Name] = source
	}

	state, readable, err := ReadState(configDir)
	if err != nil {
		readable = false
	}
	discoveryTargets := append([]string{}, opts.Targets...)
	legacyCodexRoot := filepath.Join(home, ".codex", "skills")
	if defaultTargets {
		discoveryTargets = append(discoveryTargets, legacyCodexRoot)
	}
	localSkills := listLocalSkills(discoveryTargets)
	plan := PlanSync(PlanInput{
		Version:        opts.Version,
		OfficialSkills: official,
		LocalSkills:    localSkills,
		PreviousState:  state,
		StateReadable:  readable,
		Force:          opts.Force,
	})

	result := &Result{
		Action:         "synced",
		Version:        opts.Version,
		Lang:           opts.Lang,
		Official:       nonNilStrings(plan.OfficialSkills),
		Updated:        nonNilStrings(plan.ToUpdate),
		Added:          nonNilStrings(plan.Added),
		SkippedDeleted: nonNilStrings(plan.SkippedDeleted),
		TargetDirs:     append([]string{}, opts.Targets...),
		StatePath:      StatePath(configDir),
	}

	stamp := opts.Now().UTC().Format("20060102T150405Z")
	backupRoot := filepath.Join(configDir, "backups")
	if err := os.MkdirAll(backupRoot, 0o700); err != nil {
		return result, err
	}
	backupDir, err := os.MkdirTemp(backupRoot, stamp+"-")
	if err != nil {
		return result, err
	}
	// Remove only an empty reservation. Existing recovery data is retained.
	defer os.Remove(backupDir)
	for targetIndex, target := range opts.Targets {
		for _, skillName := range plan.ToUpdate {
			source := sourceByName[skillName]
			backedUp, err := syncOneSkill(opts.SourceFS, source, target, backupDir, targetIndex)
			if backedUp {
				result.BackupCount++
			}
			if err != nil {
				return result, err
			}
		}
	}

	if defaultTargets {
		for _, name := range plan.ToUpdate {
			legacy := filepath.Join(legacyCodexRoot, name)
			canonical := filepath.Join(home, ".agents", "skills", name)
			legacyInfo, legacyErr := os.Stat(legacy)
			canonicalInfo, canonicalErr := os.Stat(canonical)
			if legacyErr == nil && canonicalErr == nil && os.SameFile(legacyInfo, canonicalInfo) {
				continue // A user-managed alias already points at the canonical tree.
			}
			backedUp, retired, err := retireLegacySkill(filepath.Join(legacyCodexRoot, name), filepath.Join(backupDir, "legacy-codex", name))
			if backedUp {
				result.BackupCount++
			}
			if retired {
				result.MigratedLegacy = append(result.MigratedLegacy, name)
			}
			if err != nil {
				return result, err
			}
		}
	}

	state = &State{
		Version:              opts.Version,
		Lang:                 opts.Lang,
		OfficialSkills:       plan.OfficialSkills,
		UpdatedSkills:        plan.ToUpdate,
		AddedOfficialSkills:  plan.Added,
		SkippedDeletedSkills: plan.SkippedDeleted,
		TargetDirs:           opts.Targets,
		UpdatedAt:            opts.Now().UTC().Format(time.RFC3339),
	}
	if err := WriteState(configDir, *state); err != nil {
		return result, err
	}
	return result, nil
}

func DefaultTargets(home string) []string {
	return []string{
		filepath.Join(home, ".agents", "skills"),
		filepath.Join(home, ".claude", "skills"),
	}
}

func StatePath(configDir string) string {
	return filepath.Join(configDir, StateFileName)
}

func ReadState(configDir string) (*State, bool, error) {
	data, err := os.ReadFile(StatePath(configDir))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, err
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, false, err
	}
	ensureStateSlices(&state)
	return &state, true, nil
}

func WriteState(configDir string, state State) error {
	ensureStateSlices(&state)
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(configDir, ".skills-state-*.json")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, StatePath(configDir))
}

func listSourceSkills(sourceFS fs.FS, lang string) ([]sourceSkill, error) {
	entries, err := fs.ReadDir(sourceFS, lang)
	if err != nil {
		return nil, err
	}
	var out []sourceSkill
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "bohrium-") {
			continue
		}
		skillPath := filepath.ToSlash(filepath.Join(lang, entry.Name()))
		data, err := fs.ReadFile(sourceFS, filepath.ToSlash(filepath.Join(skillPath, "SKILL.md")))
		if err != nil {
			return nil, err
		}
		name, err := frontmatterName(string(data))
		if err != nil {
			return nil, fmt.Errorf("%s/SKILL.md: %w", skillPath, err)
		}
		if name != entry.Name() {
			return nil, fmt.Errorf("%s: frontmatter name %q does not match directory", skillPath, name)
		}
		out = append(out, sourceSkill{Name: name, Path: skillPath})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func syncOneSkill(sourceFS fs.FS, source sourceSkill, targetRoot, backupDir string, targetIndex int) (bool, error) {
	if source.Name == "" || source.Path == "" {
		return false, errors.New("empty source skill")
	}
	if err := os.MkdirAll(targetRoot, 0o755); err != nil {
		return false, err
	}
	dest := filepath.Join(targetRoot, source.Name)
	tmp, err := os.MkdirTemp(targetRoot, "."+source.Name+".tmp-")
	if err != nil {
		return false, err
	}
	tmpActive := true
	defer func() {
		if tmpActive {
			_ = os.RemoveAll(tmp)
		}
	}()
	if err := copySkillFromFS(sourceFS, source.Path, tmp); err != nil {
		return false, err
	}

	backupMade := false
	backupPath := filepath.Join(backupDir, fmt.Sprintf("target-%d", targetIndex), source.Name)
	if _, err := os.Stat(dest); err == nil {
		if err := copyDir(dest, backupPath); err != nil {
			return false, err
		}
		backupMade = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}

	if err := publishSkillDirectory(tmp, dest, os.Rename); err != nil {
		return backupMade, err
	}
	tmpActive = false
	return backupMade, nil
}

// retireLegacySkill removes only selected official names after their complete
// canonical copies were installed. The copied backup retains user additions.
func retireLegacySkill(dest, backupPath string) (backedUp, retired bool, err error) {
	if _, err := os.Lstat(dest); errors.Is(err, os.ErrNotExist) {
		return false, false, nil
	} else if err != nil {
		return false, false, err
	}
	if err := copyDir(dest, backupPath); err != nil {
		return false, false, err
	}
	holder, err := os.MkdirTemp(filepath.Dir(dest), "."+filepath.Base(dest)+".retired-")
	if err != nil {
		return true, false, err
	}
	defer os.Remove(holder)
	previous := filepath.Join(holder, "previous")
	if err := os.Rename(dest, previous); err != nil {
		return true, false, fmt.Errorf("retire duplicate legacy Codex Skill %s (backup at %s): %w", dest, backupPath, err)
	}
	if err := os.RemoveAll(holder); err != nil {
		return true, true, fmt.Errorf("legacy Codex Skill retired and backed up at %s; cleanup failed at %s: %w", backupPath, previous, err)
	}
	return true, true, nil
}

// publishSkillDirectory never recursively deletes the active installation.
// The rename seam allows filesystem fault tests without changing Sync's API.
func publishSkillDirectory(staged, dest string, rename func(string, string) error) error {
	if _, err := os.Lstat(dest); errors.Is(err, os.ErrNotExist) {
		return rename(staged, dest)
	} else if err != nil {
		return err
	}
	holder, err := os.MkdirTemp(filepath.Dir(dest), "."+filepath.Base(dest)+".previous-")
	if err != nil {
		return err
	}
	defer os.Remove(holder) // Only removes an empty holder, never recovery files.
	previous := filepath.Join(holder, "previous")
	if err := rename(dest, previous); err != nil {
		return fmt.Errorf("preserve previous skill %s: %w", dest, err)
	}
	if publishErr := rename(staged, dest); publishErr != nil {
		if rollbackErr := rename(previous, dest); rollbackErr != nil {
			return fmt.Errorf("skill publication and rollback failed; previous installation retained at %s: %w", previous, errors.Join(publishErr, rollbackErr))
		}
		return fmt.Errorf("skill publication failed; previous installation restored: %w", publishErr)
	}
	if err := os.RemoveAll(holder); err != nil {
		return fmt.Errorf("new skill is active at %s; previous-copy cleanup failed at %s: %w", dest, previous, err)
	}
	return nil
}

func copySkillFromFS(sourceFS fs.FS, sourcePath, destRoot string) error {
	return fs.WalkDir(sourceFS, sourcePath, func(filePath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(filepath.FromSlash(sourcePath), filepath.FromSlash(filePath))
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		dest := filepath.Join(destRoot, rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		if d.IsDir() {
			return os.MkdirAll(dest, writableDirMode(info.Mode().Perm()))
		}
		data, err := fs.ReadFile(sourceFS, filePath)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		mode := info.Mode().Perm()
		if mode == 0 {
			mode = 0o644
		}
		return os.WriteFile(dest, data, mode)
	})
}

func copyDir(src, dest string) error {
	resolved, err := filepath.EvalSymlinks(src)
	if err != nil {
		return err
	}
	src = resolved
	if err := os.RemoveAll(dest); err != nil {
		return err
	}
	return filepath.WalkDir(src, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(dest, rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		if d.IsDir() {
			return os.MkdirAll(target, writableDirMode(info.Mode().Perm()))
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode().Perm())
	})
}

func listLocalSkills(targets []string) []string {
	seen := map[string]bool{}
	for _, target := range targets {
		entries, err := os.ReadDir(target)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			isDir := entry.IsDir()
			if entry.Type()&os.ModeSymlink != 0 {
				info, err := os.Stat(filepath.Join(target, entry.Name()))
				isDir = err == nil && info.IsDir()
			}
			if isDir {
				seen[entry.Name()] = true
			}
		}
	}
	return sortedKeys(seen)
}

func writableDirMode(mode os.FileMode) os.FileMode {
	if mode == 0 {
		return 0o755
	}
	return mode | 0o700
}

func resolveHome(home string) (string, error) {
	if home != "" {
		return home, nil
	}
	return os.UserHomeDir()
}

func frontmatterName(text string) (string, error) {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", fmt.Errorf("missing frontmatter")
	}
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "---" {
			break
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(key) != "name" {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if value == "" {
			return "", fmt.Errorf("empty name")
		}
		return value, nil
	}
	return "", fmt.Errorf("frontmatter name not found")
}

func ensureStateSlices(state *State) {
	if state.OfficialSkills == nil {
		state.OfficialSkills = []string{}
	}
	if state.UpdatedSkills == nil {
		state.UpdatedSkills = []string{}
	}
	if state.AddedOfficialSkills == nil {
		state.AddedOfficialSkills = []string{}
	}
	if state.SkippedDeletedSkills == nil {
		state.SkippedDeletedSkills = []string{}
	}
	if state.TargetDirs == nil {
		state.TargetDirs = []string{}
	}
}

func uniqueSorted(values []string) []string {
	return sortedKeys(toSet(values))
}

func toSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out[value] = true
		}
	}
	return out
}

func intersection(values []string, allowed map[string]bool) []string {
	out := map[string]bool{}
	for _, value := range values {
		if allowed[value] {
			out[value] = true
		}
	}
	return sortedKeys(out)
}

func sortedKeys(values map[string]bool) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
