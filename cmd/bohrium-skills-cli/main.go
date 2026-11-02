package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	bohriumskills "github.com/dptech-corp/bohrium-skills"
	"github.com/dptech-corp/bohrium-skills/internal/build"
	"github.com/dptech-corp/bohrium-skills/internal/skillsnotice"
	"github.com/dptech-corp/bohrium-skills/internal/syncer"
	"github.com/dptech-corp/bohrium-skills/internal/updatecheck"
	"github.com/dptech-corp/bohrium-skills/internal/updater"
)

const (
	refreshCacheCommand = "__refresh-update-cache"
)

type noticeState struct {
	Update *updatecheck.UpdateInfo
	Skills *skillsnotice.StaleNotice
}

var activeNoticeState struct {
	notices *noticeState
	json    bool
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}
	notices := setupNotices(args)
	activateNotices(notices, hasJSONFlag(args))
	defer activateNotices(nil, false)

	var code int
	switch args[0] {
	case refreshCacheCommand:
		code = runRefreshUpdateCache(args[1:], stderr)
	case "install":
		code = runInstall(args[1:], stdout, stderr)
	case "update":
		code = runUpdate(args[1:], stdout, stderr)
	case "status":
		code = runStatus(args[1:], stdout, stderr)
	case "list":
		code = runList(args[1:], stdout, stderr)
	case "version":
		code = runVersion(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		printUsage(stdout)
		code = 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		printUsage(stderr)
		code = 2
	}
	if code == 0 && !hasJSONFlag(args) {
		writeTextNotices(stderr, notices)
	}
	return code
}

func runInstall(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.SetOutput(stderr)
	lang := fs.String("lang", "zh", "skill language: zh or en")
	force := fs.Bool("force", false, "restore all official skills")
	jsonOut := fs.Bool("json", false, "print JSON output")
	home := fs.String("home", "", "override home directory")
	configDir := fs.String("config-dir", "", "override config directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	result, err := syncer.Sync(syncer.SyncOptions{
		Version:   build.Version,
		Lang:      *lang,
		HomeDir:   *home,
		ConfigDir: *configDir,
		Force:     *force,
		SourceFS:  bohriumskills.EmbeddedFS,
	})
	if err != nil {
		return reportError(stdout, stderr, *jsonOut, err)
	}
	if *jsonOut {
		printJSON(stdout, result)
		return 0
	}
	fmt.Fprintf(stdout, "Skills synced: %d official, %d updated, %d added, %d skipped because deleted locally\n",
		len(result.Official), len(result.Updated), len(result.Added), len(result.SkippedDeleted))
	fmt.Fprintf(stdout, "Targets: %s\n", strings.Join(result.TargetDirs, ", "))
	if result.BackupCount > 0 {
		fmt.Fprintf(stdout, "Backups written: %d\n", result.BackupCount)
	}
	return 0
}

func runUpdate(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(stderr)
	lang := fs.String("lang", "zh", "skill language: zh or en")
	force := fs.Bool("force", false, "force reinstall/update")
	check := fs.Bool("check", false, "check only")
	jsonOut := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	runner := updater.RealRunner{
		SyncFunc: func(version string, force bool) error {
			return syncForVersion(version, *lang, force)
		},
	}
	result, err := updater.Update(updater.UpdateOptions{
		CurrentVersion: build.Version,
		Force:          *force,
		Check:          *check,
		Runner:         runner,
	})
	if err != nil {
		return reportError(stdout, stderr, *jsonOut, err)
	}
	if *jsonOut {
		printJSON(stdout, result)
		return 0
	}
	fmt.Fprintln(stdout, result.Message)
	if result.Action == "manual_required" {
		fmt.Fprintf(stdout, "Download: %s\n", result.URL)
	}
	return 0
}

func runStatus(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOut := fs.Bool("json", false, "print JSON output")
	home := fs.String("home", "", "override home directory")
	configDirFlag := fs.String("config-dir", "", "override config directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	homeDir, err := resolveHome(*home)
	if err != nil {
		return reportError(stdout, stderr, *jsonOut, err)
	}
	configDir := *configDirFlag
	if configDir == "" {
		configDir = filepath.Join(homeDir, ".config", "bohrium-skills-cli")
	}
	state, readable, readErr := syncer.ReadState(configDir)
	targets := syncer.DefaultTargets(homeDir)
	if state != nil && len(state.TargetDirs) > 0 {
		targets = state.TargetDirs
	}
	status := map[string]interface{}{
		"version":        build.Version,
		"state_readable": readable,
		"state_path":     syncer.StatePath(configDir),
		"targets":        inspectTargets(targets),
	}
	if state != nil {
		status["synced_version"] = state.Version
		status["lang"] = state.Lang
		status["official"] = len(state.OfficialSkills)
		status["updated_at"] = state.UpdatedAt
	}
	if readErr != nil {
		status["state_error"] = readErr.Error()
	}
	if *jsonOut {
		printJSON(stdout, status)
		return 0
	}
	fmt.Fprintf(stdout, "bohrium-skills-cli %s\n", build.Version)
	if state != nil {
		fmt.Fprintf(stdout, "skills: version=%s lang=%s official=%d\n", state.Version, state.Lang, len(state.OfficialSkills))
	} else {
		fmt.Fprintln(stdout, "skills: not synced")
	}
	for _, target := range inspectTargets(targets) {
		fmt.Fprintf(stdout, "%s exists=%v skills=%d\n", target.Dir, target.Exists, target.SkillCount)
	}
	return 0
}

func runList(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(stderr)
	lang := fs.String("lang", "zh", "skill language: zh or en")
	jsonOut := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	skills, err := bohriumskills.OfficialSkills(*lang)
	if err != nil {
		return reportError(stdout, stderr, *jsonOut, err)
	}
	if *jsonOut {
		printJSON(stdout, map[string]interface{}{"lang": *lang, "skills": skills})
		return 0
	}
	for _, skill := range skills {
		fmt.Fprintln(stdout, skill.Name)
	}
	return 0
}

func runVersion(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOut := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *jsonOut {
		printJSON(stdout, map[string]string{"version": build.Version})
		return 0
	}
	fmt.Fprintln(stdout, build.Version)
	return 0
}

func syncForVersion(version, lang string, force bool) error {
	if normalizeVersion(version) != normalizeVersion(build.Version) {
		binary, err := installedNPMBinary(version)
		if err != nil {
			return err
		}
		args := []string{"install", "--lang", lang, "--json"}
		if force {
			args = append(args, "--force")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		out, err := installerCommandOutput(ctx, binary, args...)
		if err != nil {
			return fmt.Errorf("verified new binary skill sync failed: %w", err)
		}
		var receipt syncer.Result
		if err := json.Unmarshal(out, &receipt); err != nil {
			return fmt.Errorf("new binary returned an invalid sync receipt: %w", err)
		}
		if receipt.Action != "synced" || receipt.Lang != lang || normalizeVersion(receipt.Version) != normalizeVersion(version) {
			return fmt.Errorf("new binary sync receipt does not match requested version %s and language %s", version, lang)
		}
		return nil
	}
	_, err := syncer.Sync(syncer.SyncOptions{
		Version:  version,
		Lang:     lang,
		Force:    force,
		SourceFS: bohriumskills.EmbeddedFS,
	})
	return err
}

// installedNPMBinary follows the same npm prefix used by `npm install -g`.
// PATH may still resolve a different manual binary, so it is not authoritative.
func installedNPMBinary(version string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	out, err := installerCommandOutput(ctx, "npm", "root", "-g")
	if err != nil {
		return "", fmt.Errorf("resolve installed npm package: %w", err)
	}
	root := strings.TrimSpace(string(out))
	if !filepath.IsAbs(root) {
		return "", fmt.Errorf("npm root -g did not return an absolute package directory")
	}
	packageDir := filepath.Join(root, updater.NpmPackage)
	data, err := os.ReadFile(filepath.Join(packageDir, "package.json"))
	if err != nil {
		return "", fmt.Errorf("read installed npm package identity: %w", err)
	}
	var pkg struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", fmt.Errorf("read installed npm package identity: %w", err)
	}
	if pkg.Name != updater.NpmPackage || normalizeVersion(pkg.Version) != normalizeVersion(version) {
		return "", fmt.Errorf("installed npm package identity/version does not match %s@%s", updater.NpmPackage, version)
	}
	binary := filepath.Join(packageDir, "bin", updater.NpmPackage)
	if _, err := os.Stat(binary); os.IsNotExist(err) {
		binary += ".exe"
	}
	out, err = installerCommandOutput(ctx, binary, "version", "--json")
	if err != nil {
		return "", fmt.Errorf("verify installed npm binary version: %w", err)
	}
	var reported struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(out, &reported); err != nil {
		return "", fmt.Errorf("installed binary returned an invalid version receipt: %w", err)
	}
	if normalizeVersion(reported.Version) != normalizeVersion(version) {
		return "", fmt.Errorf("installed binary reports version %s, expected %s; no skill sync was started", reported.Version, version)
	}
	return binary, nil
}

func installerCommandOutput(ctx context.Context, binary string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, binary, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return out, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()+"\n"+string(out)))
	}
	return out, nil
}

type targetStatus struct {
	Dir        string `json:"dir"`
	Exists     bool   `json:"exists"`
	SkillCount int    `json:"skill_count"`
}

func inspectTargets(targets []string) []targetStatus {
	statuses := make([]targetStatus, 0, len(targets))
	for _, target := range targets {
		status := targetStatus{Dir: target}
		entries, err := os.ReadDir(target)
		if err == nil {
			status.Exists = true
			for _, entry := range entries {
				if entry.IsDir() && strings.HasPrefix(entry.Name(), "bohrium-") {
					status.SkillCount++
				}
			}
		}
		statuses = append(statuses, status)
	}
	return statuses
}

func resolveHome(home string) (string, error) {
	if home != "" {
		return home, nil
	}
	return os.UserHomeDir()
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "v")
	return strings.TrimPrefix(version, "V")
}

func setupNotices(args []string) *noticeState {
	updatecheck.SetPending(nil)
	skillsnotice.SetPending(nil)
	if skipNotices(args) {
		return nil
	}

	if update := updatecheck.CheckCached(build.Version); update != nil {
		updatecheck.SetPending(update)
	}
	startUpdateCacheRefresh(build.Version)

	configDir := noticeConfigDir(args)
	if skills := skillsnotice.Check(build.Version, configDir); skills != nil {
		skillsnotice.SetPending(skills)
	}
	return currentNotices()
}

func currentNotices() *noticeState {
	notices := &noticeState{
		Update: updatecheck.GetPending(),
		Skills: skillsnotice.GetPending(),
	}
	if notices.Update == nil && notices.Skills == nil {
		return nil
	}
	return notices
}

func activateNotices(notices *noticeState, jsonOut bool) {
	activeNoticeState.notices = notices
	activeNoticeState.json = jsonOut
}

func skipNotices(args []string) bool {
	if len(args) == 0 {
		return true
	}
	switch args[0] {
	case refreshCacheCommand, "help", "-h", "--help", "version", "completion", "completions":
		return true
	}
	for _, arg := range args {
		switch arg {
		case "-h", "--help", "help":
			return true
		}
		if strings.Contains(arg, "completion") {
			return true
		}
	}
	return false
}

func hasJSONFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--json" || strings.HasPrefix(arg, "--json=") {
			return true
		}
	}
	return false
}

func noticeConfigDir(args []string) string {
	var home string
	var configDir string
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--home" && i+1 < len(args):
			home = args[i+1]
			i++
		case strings.HasPrefix(arg, "--home="):
			home = strings.TrimPrefix(arg, "--home=")
		case arg == "--config-dir" && i+1 < len(args):
			configDir = args[i+1]
			i++
		case strings.HasPrefix(arg, "--config-dir="):
			configDir = strings.TrimPrefix(arg, "--config-dir=")
		}
	}
	if configDir != "" {
		return configDir
	}
	if home == "" {
		resolved, err := os.UserHomeDir()
		if err == nil {
			home = resolved
		}
	}
	if home == "" {
		return filepath.Join(".", ".config", "bohrium-skills-cli")
	}
	return filepath.Join(home, ".config", "bohrium-skills-cli")
}

func writeTextNotices(stderr io.Writer, notices *noticeState) {
	if notices == nil {
		return
	}
	if notices.Update != nil {
		fmt.Fprintln(stderr, notices.Update.Message())
	}
	if notices.Skills != nil {
		fmt.Fprintln(stderr, notices.Skills.Message())
	}
}

func startUpdateCacheRefresh(version string) {
	if !updatecheck.NeedsRefresh(version) {
		return
	}
	exe, err := os.Executable()
	if err != nil {
		return
	}
	cmd := exec.Command(exe, refreshCacheCommand, "--version", version)
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return
	}
	_ = cmd.Process.Release()
}

func runRefreshUpdateCache(args []string, stderr io.Writer) int {
	fs := flag.NewFlagSet(refreshCacheCommand, flag.ContinueOnError)
	fs.SetOutput(stderr)
	version := fs.String("version", build.Version, "current version")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	updatecheck.RefreshCache(*version)
	return 0
}

func noticeObject(notices *noticeState) map[string]interface{} {
	if notices == nil {
		return nil
	}
	out := make(map[string]interface{})
	if notices.Update != nil {
		out["update"] = map[string]string{
			"current": notices.Update.Current,
			"latest":  notices.Update.Latest,
			"message": notices.Update.Message(),
			"command": "bohrium-skills-cli update",
		}
	}
	if notices.Skills != nil {
		out["skills"] = map[string]string{
			"current": notices.Skills.Current,
			"target":  notices.Skills.Target,
			"message": notices.Skills.Message(),
			"command": "bohrium-skills-cli update",
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func reportError(stdout, stderr io.Writer, jsonOut bool, err error) int {
	if jsonOut {
		printJSONRaw(stdout, map[string]interface{}{
			"ok": false,
			"error": map[string]string{
				"message": err.Error(),
			},
		})
		return 1
	}
	fmt.Fprintln(stderr, "error:", err)
	return 1
}

func printJSON(w io.Writer, value interface{}) {
	if activeNoticeState.json {
		value = withNotice(value, noticeObject(activeNoticeState.notices))
	}
	printJSONRaw(w, value)
}

func printJSONRaw(w io.Writer, value interface{}) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(value)
}

func withNotice(value interface{}, notice map[string]interface{}) interface{} {
	if len(notice) == 0 {
		return value
	}
	data, err := json.Marshal(value)
	if err != nil {
		return value
	}
	var object map[string]interface{}
	if err := json.Unmarshal(data, &object); err != nil || object == nil {
		return value
	}
	object["_notice"] = notice
	return object
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `bohrium-skills-cli manages bundled Bohrium AI skills.

Usage:
  bohrium-skills-cli install [--lang zh|en] [--force] [--json]
  bohrium-skills-cli update [--check] [--force] [--json]
  bohrium-skills-cli status [--json]
  bohrium-skills-cli list [--lang zh|en] [--json]
  bohrium-skills-cli version [--json]`)
}
