package updater

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dptech-corp/bohrium-skills/internal/updatecheck"
)

const (
	NpmPackage = "bohrium-skills-cli"
	RepoURL    = "https://github.com/dptech-corp/bohrium-skills"
)

type InstallMethod string

const (
	InstallNPM    InstallMethod = "npm"
	InstallManual InstallMethod = "manual"
)

type SyncRequest struct {
	Version       string
	Force         bool
	UpdatedBinary bool
}

type Runner interface {
	FetchLatest() (string, error)
	DetectInstallMethod() InstallMethod
	Install(version string) error
	Sync(request SyncRequest) error
}

type UpdateOptions struct {
	CurrentVersion string
	Force          bool
	Check          bool
	Runner         Runner
}

type Result struct {
	Action          string        `json:"action"`
	PreviousVersion string        `json:"previous_version"`
	CurrentVersion  string        `json:"current_version"`
	LatestVersion   string        `json:"latest_version"`
	InstallMethod   InstallMethod `json:"install_method"`
	URL             string        `json:"url,omitempty"`
	Message         string        `json:"message"`
	SkillsAction    string        `json:"skills_action,omitempty"`
}

type RealRunner struct {
	SyncFunc func(request SyncRequest) error
	Timeout  time.Duration
}

type FakeRunner struct {
	LatestVersion string
	LatestErr     error
	InstallFunc   func(version string) error
	SyncFunc      func(request SyncRequest) error
	InstallMethod InstallMethod
}

func Update(opts UpdateOptions) (*Result, error) {
	if opts.Runner == nil {
		return nil, errors.New("update runner is nil")
	}
	cur := normalizeVersion(opts.CurrentVersion)
	latestRaw, err := opts.Runner.FetchLatest()
	if err != nil {
		return nil, err
	}
	latest := normalizeVersion(latestRaw)
	method := opts.Runner.DetectInstallMethod()
	result := &Result{
		PreviousVersion: cur,
		CurrentVersion:  cur,
		LatestVersion:   latest,
		InstallMethod:   method,
		URL:             RepoURL + "/releases/tag/v" + strings.TrimPrefix(latest, "v"),
	}

	newer := updatecheck.IsNewer(latest, cur)
	if opts.Check {
		if newer {
			result.Action = "update_available"
			result.Message = fmt.Sprintf("bohrium-skills-cli %s -> %s available", cur, latest)
		} else {
			result.Action = "already_up_to_date"
			result.Message = fmt.Sprintf("bohrium-skills-cli %s is already up to date", cur)
		}
		return result, nil
	}

	if method != InstallNPM {
		if err := opts.Runner.Sync(SyncRequest{Version: cur, Force: opts.Force}); err != nil {
			return result, err
		}
		result.SkillsAction = "synced"
		if newer {
			result.Action = "manual_required"
			result.Message = "automatic binary update unavailable; synced skills from the current binary"
		} else {
			result.Action = "already_up_to_date"
			result.Message = "binary already up to date; synced skills"
		}
		return result, nil
	}

	targetVersion := cur
	if newer || opts.Force {
		if err := opts.Runner.Install(latest); err != nil {
			return result, err
		}
		targetVersion = latest

	} else {
		result.Action = "already_up_to_date"
		result.Message = fmt.Sprintf("bohrium-skills-cli %s is already up to date", cur)
	}
	if err := opts.Runner.Sync(SyncRequest{Version: targetVersion, Force: opts.Force, UpdatedBinary: newer || opts.Force}); err != nil {
		return result, err
	}
	if newer || opts.Force {
		result.Action = "updated"
		result.CurrentVersion = latest
		result.Message = fmt.Sprintf("bohrium-skills-cli updated from %s to %s", cur, latest)
	}
	result.SkillsAction = "synced"
	return result, nil
}

func (r RealRunner) FetchLatest() (string, error) {
	return updatecheck.FetchLatest()
}

func (r RealRunner) DetectInstallMethod() InstallMethod {
	exe, err := os.Executable()
	if err != nil {
		return InstallManual
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		resolved = exe
	}
	binDir := filepath.Dir(resolved)
	packageDir := filepath.Dir(binDir)
	if filepath.Base(binDir) != "bin" || filepath.Base(packageDir) != NpmPackage || filepath.Base(filepath.Dir(packageDir)) != "node_modules" {
		return InstallManual
	}
	data, err := os.ReadFile(filepath.Join(packageDir, "package.json"))
	if err != nil {
		return InstallManual
	}
	var pkg struct {
		Name string `json:"name"`
	}
	if json.Unmarshal(data, &pkg) != nil || pkg.Name != NpmPackage {
		return InstallManual
	}
	return InstallNPM
}

func (r RealRunner) Install(version string) error {
	timeout := r.Timeout
	if timeout == 0 {
		timeout = 10 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "npm", "install", "-g", NpmPackage+"@"+version)
	// The parent verifies the installed binary before the one explicit sync.
	cmd.Env = append(os.Environ(), "BOHRIUM_SKILLS_CLI_NO_POSTINSTALL_SYNC=1")
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm install failed: %w\n%s", err, strings.TrimSpace(combined.String()))
	}
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("npm install timed out after %s", timeout)
	}
	return nil
}

func (r RealRunner) Sync(request SyncRequest) error {
	if r.SyncFunc == nil {
		return errors.New("sync function is nil")
	}
	return r.SyncFunc(request)
}

func (f FakeRunner) FetchLatest() (string, error) {
	if f.LatestErr != nil {
		return "", f.LatestErr
	}
	return f.LatestVersion, nil
}

func (f FakeRunner) DetectInstallMethod() InstallMethod {
	if f.InstallMethod == "" {
		return InstallManual
	}
	return f.InstallMethod
}

func (f FakeRunner) Install(version string) error {
	if f.InstallFunc == nil {
		return nil
	}
	return f.InstallFunc(version)
}

func (f FakeRunner) Sync(request SyncRequest) error {
	if f.SyncFunc == nil {
		return nil
	}
	return f.SyncFunc(request)
}

func IsNewer(candidate, current string) bool {
	candidateParts := versionParts(candidate)
	currentParts := versionParts(current)
	for i := 0; i < len(candidateParts) || i < len(currentParts); i++ {
		var c, cur int
		if i < len(candidateParts) {
			c = candidateParts[i]
		}
		if i < len(currentParts) {
			cur = currentParts[i]
		}
		if c > cur {
			return true
		}
		if c < cur {
			return false
		}
	}
	return false
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "v")
	version = strings.TrimPrefix(version, "V")
	return version
}

func versionParts(version string) []int {
	version = normalizeVersion(version)
	if index := strings.IndexAny(version, "+-"); index >= 0 {
		version = version[:index]
	}
	raw := strings.Split(version, ".")
	parts := make([]int, 0, len(raw))
	for _, part := range raw {
		n, err := strconv.Atoi(part)
		if err != nil {
			parts = append(parts, 0)
			continue
		}
		parts = append(parts, n)
	}
	return parts
}
