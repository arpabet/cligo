/*
 * Copyright (c) 2026 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import (
	"strings"
	"testing"

	"go.arpabet.com/glue"
)

// ─── Config file loading ─────────────────────────────────────────────────────

func TestConfigFile_PropertiesFormat(t *testing.T) {
	path := writeTempFile(t, "config.properties", `
# Java properties format
app.profile = production
app.port = 443
`)
	cmd := &propCmd{}
	withArgs([]string{"app", "propcmd"}, func() {
		if err := Run(ConfigFile(path), Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "production" {
		t.Errorf("expected Profile=production, got %q", cmd.Profile)
	}
}

func TestConfigFile_YAMLFormat(t *testing.T) {
	path := writeTempFile(t, "config.yaml", `
app:
  profile: dev
  port: "8080"
`)
	cmd := &propCmd{}
	withArgs([]string{"app", "propcmd"}, func() {
		if err := Run(ConfigFile(path), Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "dev" {
		t.Errorf("expected Profile=dev, got %q", cmd.Profile)
	}
	if cmd.Port != "8080" {
		t.Errorf("expected Port=8080, got %q", cmd.Port)
	}
}

func TestConfigFile_YAMLFormat_YmlExtension(t *testing.T) {
	path := writeTempFile(t, "config.yml", `
app:
  profile: test
  port: "5000"
`)
	cmd := &propCmd{}
	withArgs([]string{"app", "propcmd"}, func() {
		if err := Run(ConfigFile(path), Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "test" {
		t.Errorf("expected Profile=test, got %q", cmd.Profile)
	}
}

func TestConfigFile_JSONFormat(t *testing.T) {
	path := writeTempFile(t, "config.json", `{
  "app": {
    "profile": "live",
    "port": "443"
  }
}`)
	cmd := &propCmd{}
	withArgs([]string{"app", "propcmd"}, func() {
		if err := Run(ConfigFile(path), Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "live" {
		t.Errorf("expected Profile=live, got %q", cmd.Profile)
	}
	if cmd.Port != "443" {
		t.Errorf("expected Port=443, got %q", cmd.Port)
	}
}

func TestConfigFile_FirstExistingFileWins(t *testing.T) {
	path := writeTempFile(t, "app.properties", "app.profile=from-props\napp.port=80")
	cmd := &propCmd{}
	withArgs([]string{"app", "propcmd"}, func() {
		if err := Run(ConfigFile("/nonexistent/app.properties"), ConfigFile(path), Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "from-props" {
		t.Errorf("expected Profile=from-props, got %q", cmd.Profile)
	}
}

func TestConfigFile_NoFileFound_NotAnError(t *testing.T) {
	withArgs([]string{"app", "--help"}, func() {
		captureOutput(func() {
			err := Run(ConfigFile("/nonexistent/a.properties"), ConfigFile("/nonexistent/b.yaml"))
			if err != nil {
				t.Errorf("missing config files should not be an error, got: %v", err)
			}
		})
	})
}

func TestConfigFile_MergesWithExistingProperties(t *testing.T) {
	path := writeTempFile(t, "app.properties", `app.port=7070`)
	props := glue.NewProperties()
	props.Set("app.profile", "preset")

	cmd := &propCmd{}
	withArgs([]string{"app", "propcmd"}, func() {
		if err := Run(Properties(props), ConfigFile(path), Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "preset" {
		t.Errorf("expected Profile=preset (from existing props), got %q", cmd.Profile)
	}
	if cmd.Port != "7070" {
		t.Errorf("expected Port=7070 (from config file), got %q", cmd.Port)
	}
}

func TestConfigFile_UnsupportedExtension_ReturnsError(t *testing.T) {
	path := writeTempFile(t, "config.xml", `<config/>`)
	withArgs([]string{"app", "--help"}, func() {
		captureOutput(func() {
			err := Run(ConfigFile(path))
			if err == nil {
				t.Error("expected error for unsupported config format")
			}
			if !strings.Contains(err.Error(), "unsupported") {
				t.Errorf("expected 'unsupported' in error, got: %v", err)
			}
		})
	})
}

// ─── --config flag tests ─────────────────────────────────────────────────────

func TestConfigFlag_CLIFlag(t *testing.T) {
	path := writeTempFile(t, "config.properties", "app.profile=fromflag\napp.port=1234")
	cmd := &propCmd{}
	withArgs([]string{"app", "--config", path, "propcmd"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "fromflag" {
		t.Errorf("expected Profile=fromflag, got %q", cmd.Profile)
	}
	if cmd.Port != "1234" {
		t.Errorf("expected Port=1234, got %q", cmd.Port)
	}
}

func TestConfigFlag_CLIFlag_EqualsForm(t *testing.T) {
	path := writeTempFile(t, "config.properties", "app.profile=eqform\napp.port=5678")
	cmd := &propCmd{}
	withArgs([]string{"app", "--config=" + path, "propcmd"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "eqform" {
		t.Errorf("expected Profile=eqform, got %q", cmd.Profile)
	}
}

func TestConfigFlag_Repeated(t *testing.T) {
	// First file exists, second is nonexistent — first wins
	path := writeTempFile(t, "first.properties", "app.profile=first\napp.port=1111")
	cmd := &propCmd{}
	withArgs([]string{"app", "--config", "/nonexistent/second.properties", "--config", path, "propcmd"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "first" {
		t.Errorf("expected Profile=first, got %q", cmd.Profile)
	}
}

func TestConfigFlag_MergesWithOption(t *testing.T) {
	// ConfigFile option provides one path, --config provides another — both are candidates
	path := writeTempFile(t, "flag.properties", "app.profile=flagval\napp.port=9999")
	cmd := &propCmd{}
	withArgs([]string{"app", "--config", path, "propcmd"}, func() {
		if err := Run(ConfigFile("/nonexistent/option.properties"), Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "flagval" {
		t.Errorf("expected Profile=flagval, got %q", cmd.Profile)
	}
}

func TestConfigFlag_ShowsInHelp(t *testing.T) {
	withArgs([]string{"app", "--help"}, func() {
		out := captureOutput(func() {
			_ = Run()
		})
		if !strings.Contains(out, "--config") {
			t.Errorf("expected --config in help output, got:\n%s", out)
		}
	})
}

func TestConfigFlag_ShortFlag(t *testing.T) {
	path := writeTempFile(t, "config.properties", "app.profile=short\napp.port=4321")
	cmd := &propCmd{}
	withArgs([]string{"app", "-c", path, "propcmd"}, func() {
		if err := Run(Beans(cmd)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if cmd.Profile != "short" {
		t.Errorf("expected Profile=short, got %q", cmd.Profile)
	}
}
