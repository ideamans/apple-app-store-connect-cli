package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/pelletier/go-toml/v2"
)

const dirName = "apple-app-store-connect"

// Profile is a named set of App Store Connect API credentials.
type Profile struct {
	IssuerID string `toml:"issuer_id"`
	KeyID    string `toml:"key_id,omitempty"`
	// PrivateKey is a path to the .p8 file, absolute or relative to the config directory.
	PrivateKey string `toml:"private_key"`
}

type Config struct {
	DefaultProfile string             `toml:"default_profile,omitempty"`
	Profiles       map[string]Profile `toml:"profiles,omitempty"`
}

// Dir returns the configuration directory (~/.config/apple-app-store-connect,
// honoring XDG_CONFIG_HOME).
func Dir() (string, error) {
	if base := os.Getenv("XDG_CONFIG_HOME"); base != "" {
		return filepath.Join(base, dirName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", dirName), nil
}

func FilePath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

func KeysDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "keys"), nil
}

// Load reads config.toml. A missing file yields an empty config.
func Load() (*Config, error) {
	path, err := FilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Config{Profiles: map[string]Profile{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]Profile{}
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := toml.Marshal(c)
	if err != nil {
		return err
	}
	path, err := FilePath()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func (c *Config) ProfileNames() []string {
	names := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

var keyIDPattern = regexp.MustCompile(`^AuthKey_([A-Z0-9]+)\.p8$`)

// KeyIDFromFilename extracts the key ID from an AuthKey_XXXXXXXXXX.p8 filename.
// Returns "" if the filename does not match.
func KeyIDFromFilename(path string) string {
	m := keyIDPattern.FindStringSubmatch(filepath.Base(path))
	if m == nil {
		return ""
	}
	return m[1]
}

// Credentials is a fully resolved set of credentials ready for signing.
type Credentials struct {
	Source        string // "env" or "profile:<name>"
	IssuerID      string
	KeyID         string
	PrivateKeyPEM []byte
}

// Resolve resolves credentials with the following precedence:
//  1. Environment variables ASC_ISSUER_ID + ASC_PRIVATE_KEY_PATH (or ASC_PRIVATE_KEY_BASE64)
//  2. The profile given by the --profile flag, then ASC_PROFILE, then default_profile
func Resolve(profileFlag string) (*Credentials, error) {
	issuer := os.Getenv("ASC_ISSUER_ID")
	keyPath := os.Getenv("ASC_PRIVATE_KEY_PATH")
	keyB64 := os.Getenv("ASC_PRIVATE_KEY_BASE64")
	keyID := os.Getenv("ASC_KEY_ID")

	if issuer != "" && (keyPath != "" || keyB64 != "") {
		var pemData []byte
		switch {
		case keyB64 != "":
			if keyID == "" {
				return nil, errors.New("ASC_KEY_ID is required when using ASC_PRIVATE_KEY_BASE64")
			}
			decoded, err := base64.StdEncoding.DecodeString(keyB64)
			if err != nil {
				return nil, fmt.Errorf("decode ASC_PRIVATE_KEY_BASE64: %w", err)
			}
			pemData = decoded
		default:
			data, err := os.ReadFile(keyPath)
			if err != nil {
				return nil, fmt.Errorf("read ASC_PRIVATE_KEY_PATH: %w", err)
			}
			pemData = data
			if keyID == "" {
				keyID = KeyIDFromFilename(keyPath)
			}
			if keyID == "" {
				return nil, errors.New("cannot derive key ID from ASC_PRIVATE_KEY_PATH filename; set ASC_KEY_ID")
			}
		}
		return &Credentials{Source: "env", IssuerID: issuer, KeyID: keyID, PrivateKeyPEM: pemData}, nil
	}

	cfg, err := Load()
	if err != nil {
		return nil, err
	}
	name := profileFlag
	if name == "" {
		name = os.Getenv("ASC_PROFILE")
	}
	if name == "" {
		name = cfg.DefaultProfile
	}
	if name == "" {
		return nil, errors.New(`no profile configured; run "asc configure --issuer-id <ID> --key <AuthKey_XXX.p8>" first`)
	}
	profile, ok := cfg.Profiles[name]
	if !ok {
		path, _ := FilePath()
		return nil, fmt.Errorf("profile %q not found in %s", name, path)
	}

	resolvedPath := profile.PrivateKey
	if !filepath.IsAbs(resolvedPath) {
		dir, err := Dir()
		if err != nil {
			return nil, err
		}
		resolvedPath = filepath.Join(dir, resolvedPath)
	}
	pemData, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("read private key for profile %q: %w", name, err)
	}
	resolvedKeyID := profile.KeyID
	if resolvedKeyID == "" {
		resolvedKeyID = KeyIDFromFilename(resolvedPath)
	}
	if resolvedKeyID == "" {
		return nil, fmt.Errorf("profile %q: cannot derive key ID from %s; set key_id in config.toml", name, resolvedPath)
	}
	return &Credentials{
		Source:        "profile:" + name,
		IssuerID:      profile.IssuerID,
		KeyID:         resolvedKeyID,
		PrivateKeyPEM: pemData,
	}, nil
}
