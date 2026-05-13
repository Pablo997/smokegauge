package config

import (
	"errors"
	"fmt"
	time2 "time"
)

type Defaults struct {
	Timeout     string
	Concurrency int
}

type Check struct {
	Name       string
	Method     string
	URL        string `yaml:"url"`
	WantStatus int    `yaml:"want_status"`
}

type Config struct {
	// Version is the config file schema revision (e.g. 1), not the smokegauge binary semver.
	Version  int
	Defaults Defaults
	Checks   []Check
}

// Validate returns all independent issues found; nil or empty slice means the config is structurally acceptable.
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
