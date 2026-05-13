// Package config defines the smokegauge YAML configuration model and validation rules.
package config

import (
	"errors"
	"fmt"
	time2 "time"
)

// Defaults holds per-run settings from the config file's defaults section.
type Defaults struct {
	Timeout     string
	Concurrency int
}

// Check is a single HTTP probe described in YAML.
type Check struct {
	Name       string
	Method     string
	URL        string `yaml:"url"`
	WantStatus int    `yaml:"want_status"`
}

// Config is the root document loaded from a checks YAML file.
type Config struct {
	// Version is the config file schema revision (e.g. 1), not the smokegauge binary semver.
	Version  int
	Defaults Defaults
	Checks   []Check
}

// Validate reports problems with c. An empty return value means c is acceptable for execution.
func (c *Config) Validate() []error {
	var errs []error
	if c.Version == 0 {
		errs = append(errs, errors.New("config schema version is missing (add version: 1)"))
	} else if c.Version != 1 {
		errs = append(errs, fmt.Errorf("unsupported schema version: %d (only 1 supported)", c.Version))
	}

	_, err := time2.ParseDuration(c.Defaults.Timeout)
	if err != nil {
		errs = append(errs, err)
	}

	if c.Defaults.Concurrency <= 0 {
		errs = append(errs, fmt.Errorf("value %d is not allowed for concurrency. Please, introduce a value higher than 0", c.Defaults.Concurrency))
	}

	if len(c.Checks) == 0 {
		errs = append(errs, errors.New("check section is empty"))
	} else {
		for i, check := range c.Checks {
			if check.URL == "" {
				errs = append(errs, fmt.Errorf("the URL for check at index %d is empty", i))
			}
		}
	}
	return errs
}
