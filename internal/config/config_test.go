package config

import (
	"path/filepath"
	"testing"
)

func TestKeyIDFromFilename(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"AuthKey_AKSKY6N6D9.p8", "AKSKY6N6D9"},
		{"/some/dir/AuthKey_ABC123DEF4.p8", "ABC123DEF4"},
		{"key.p8", ""},
		{"AuthKey_lowercase.p8", ""},
		{"AuthKey_ABC123DEF4.pem", ""},
	}
	for _, c := range cases {
		if got := KeyIDFromFilename(c.path); got != c.want {
			t.Errorf("KeyIDFromFilename(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

func TestConfigSaveLoadRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load empty: %v", err)
	}
	if len(cfg.Profiles) != 0 || cfg.DefaultProfile != "" {
		t.Fatalf("expected empty config, got %+v", cfg)
	}

	cfg.DefaultProfile = "default"
	cfg.Profiles["default"] = Profile{
		IssuerID:   "issuer-uuid",
		KeyID:      "ABC123DEF4",
		PrivateKey: filepath.Join("keys", "AuthKey_ABC123DEF4.p8"),
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load after save: %v", err)
	}
	if loaded.DefaultProfile != "default" {
		t.Errorf("DefaultProfile = %q, want %q", loaded.DefaultProfile, "default")
	}
	p, ok := loaded.Profiles["default"]
	if !ok {
		t.Fatalf("profile %q missing after round trip", "default")
	}
	if p != cfg.Profiles["default"] {
		t.Errorf("profile round trip mismatch: got %+v, want %+v", p, cfg.Profiles["default"])
	}
}

func TestResolveProfileNotFound(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("ASC_ISSUER_ID", "")
	t.Setenv("ASC_PROFILE", "")

	if _, err := Resolve("missing"); err == nil {
		t.Error("Resolve with unknown profile should fail")
	}
	if _, err := Resolve(""); err == nil {
		t.Error("Resolve with no profiles configured should fail")
	}
}
