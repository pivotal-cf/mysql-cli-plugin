package multisite

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cli/cf/configuration/confighelpers"
)

// TODO: There are various testing gaps at the corner cases where interacting with the filesystem could fail in strange ways
//       We could test these corner by plumbing in a filesystem abstraction of some sort.
//       It is hard to make those corner cases fail on a "real" system, otherwise.

type Config struct {
	Dir string
}

type Target struct {
	Name         string
	Organization string
	Space        string
	API          string
}

func (t *Target) Validate() error {
	var missingFields []string

	if t.API == "" {
		missingFields = append(missingFields, "API endpoint")
	}
	if t.Organization == "" {
		missingFields = append(missingFields, "Organization")
	}
	if t.Space == "" {
		missingFields = append(missingFields, "Space")
	}

	if len(missingFields) != 0 {
		return fmt.Errorf("missing fields: [%s]", strings.Join(missingFields, ","))
	}

	return nil
}

func (t *Target) UnmarshalJSON(b []byte) error {
	var data struct {
		OrganizationFields struct {
			Name string
		} `json:"OrganizationFields"`
		SpaceFields struct {
			Name string
		} `json:"SpaceFields"`
		Target string `json:"Target"`
	}

	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	t.API = data.Target
	t.Organization = data.OrganizationFields.Name
	t.Space = data.SpaceFields.Name

	return nil
}

func NewConfig() Config {
	return Config{
		Dir: filepath.Join(confighelpers.PluginRepoDir(), ".cf", ".mysql-tools"),
	}
}

func (c Config) ListConfigs() (targets []Target, errs error) {
	files, err := filepath.Glob(filepath.Join(c.Dir, "*", ".cf", "config.json"))
	if err != nil {
		return nil, err
	}

	for _, configFilePath := range files {
		configName := filepath.Base(filepath.Dir(filepath.Dir(configFilePath)))
		t := Target{Name: configName}
		if err = c.parseTargetFromPath(configFilePath, &t); err != nil {
			errs = errors.Join(errs, err)
			return nil, err
		}
		targets = append(targets, t)
	}

	return targets, errs
}

func (Config) parseTargetFromPath(path string, t *Target) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return json.NewDecoder(f).Decode(t)
}

func (c Config) SaveConfig(path, name string) (Target, error) {
	f, err := os.Open(path)
	if err != nil {
		return Target{}, err
	}
	defer func() { _ = f.Close() }()

	t := Target{Name: name}
	if err := json.NewDecoder(f).Decode(&t); err != nil {
		return Target{}, err
	}

	if err := t.Validate(); err != nil {
		return Target{}, fmt.Errorf("saved configuration must target Cloudfoundry: %s", err)
	}

	targetConfig := filepath.Join(c.Dir, name, ".cf", "config.json")

	// what permissions should we put on the dir?
	// should we be testing this?
	if err := os.MkdirAll(filepath.Dir(targetConfig), 0700); err != nil {
		return Target{}, err
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return Target{}, fmt.Errorf("failed to reset offset in source file %s: %s", f.Name(), err)
	}

	w, err := os.OpenFile(targetConfig, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return Target{}, err
	}
	defer func() { _ = w.Close() }()

	if _, err := io.Copy(w, f); err != nil {
		return Target{}, err
	}

	if err := w.Close(); err != nil {
		return Target{}, err
	}

	return t, nil
}

func (c Config) RemoveConfig(name string) error {
	path := filepath.Join(c.Dir, name)

	if filepath.Dir(path) != c.Dir || filepath.Base(path) != name {
		return fmt.Errorf("invalid target name %q", name)
	}

	if err := os.RemoveAll(path); err != nil {
		return err
	}

	return nil
}

func (c Config) ConfigDir(name string) string {
	return filepath.Join(c.Dir, name)
}
