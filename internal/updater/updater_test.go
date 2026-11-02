package updater

import (
	"errors"
	"reflect"
	"testing"
)

func TestUpdateNPMInstallThenSyncsSkills(t *testing.T) {
	var calls []string
	runner := FakeRunner{
		LatestVersion: "1.2.0",
		InstallFunc: func(version string) error {
			calls = append(calls, "install:"+version)
			return nil
		},
		SyncFunc: func(request SyncRequest) error {
			if !request.UpdatedBinary {
				t.Fatal("installed binary was not selected for sync")
			}
			calls = append(calls, "sync:"+request.Version)
			if request.Force {
				calls = append(calls, "force")
			}
			return nil
		},
		InstallMethod: InstallNPM,
	}

	result, err := Update(UpdateOptions{
		CurrentVersion: "1.1.0",
		Force:          true,
		Runner:         runner,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if result.Action != "updated" {
		t.Fatalf("Action = %q, want updated", result.Action)
	}
	want := []string{"install:1.2.0", "sync:1.2.0", "force"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}

func TestUpdateCheckDoesNotInstallOrSync(t *testing.T) {
	called := false
	runner := FakeRunner{
		LatestVersion: "1.2.0",
		InstallFunc: func(version string) error {
			called = true
			return nil
		},
		SyncFunc: func(request SyncRequest) error {
			called = true
			return nil
		},
		InstallMethod: InstallNPM,
	}

	result, err := Update(UpdateOptions{
		CurrentVersion: "1.1.0",
		Check:          true,
		Runner:         runner,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if result.Action != "update_available" {
		t.Fatalf("Action = %q, want update_available", result.Action)
	}
	if called {
		t.Fatal("check mode installed or synced")
	}
}

func TestFailedSkillSyncDoesNotReportSuccessfulUpdate(t *testing.T) {
	failure := errors.New("new binary verification failed")
	result, err := Update(UpdateOptions{CurrentVersion: "1.0.0", Runner: FakeRunner{
		LatestVersion: "2.0.0", InstallMethod: InstallNPM,
		SyncFunc: func(SyncRequest) error { return failure },
	}})
	if !errors.Is(err, failure) {
		t.Fatalf("sync failure was lost: %v", err)
	}
	if result.Action == "updated" || result.CurrentVersion != "1.0.0" || result.SkillsAction == "synced" {
		t.Fatalf("failed sync reported success: %+v", result)
	}
}

func TestSameVersionForceStillSelectsInstalledBinary(t *testing.T) {
	installed, synced := false, false
	_, err := Update(UpdateOptions{CurrentVersion: "1.0.0", Force: true, Runner: FakeRunner{
		LatestVersion: "1.0.0", InstallMethod: InstallNPM,
		InstallFunc: func(string) error { installed = true; return nil },
		SyncFunc: func(request SyncRequest) error {
			synced = true
			if !request.UpdatedBinary || !request.Force {
				t.Fatalf("reinstalled binary not selected: %+v", request)
			}
			return nil
		},
	}})
	if err != nil || !installed || !synced {
		t.Fatalf("same-version force did not reinstall/sync: %v", err)
	}
}

func TestManualForceUsesCurrentEmbeddedSnapshot(t *testing.T) {
	_, err := Update(UpdateOptions{CurrentVersion: "1.0.0", Force: true, Runner: FakeRunner{
		LatestVersion: "2.0.0", InstallMethod: InstallManual,
		InstallFunc: func(string) error { t.Fatal("manual install invoked npm"); return nil },
		SyncFunc: func(request SyncRequest) error {
			if request.UpdatedBinary || request.Version != "1.0.0" || !request.Force {
				t.Fatalf("incorrect manual sync: %+v", request)
			}
			return nil
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
}
