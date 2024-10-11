package pgpkg

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"io"
)

// Load the settings
type configType struct {
	Package    string
	Schema     string `toml:",omitempty"` // for backward compatibility only
	Schemas    []string
	Extensions []string
	Uses       []string
	Migrations []string
}

// Read a configuration TOML file and update the package accordingly.
// If the package is already configured, it's an error.
func parseConfig(reader io.Reader) (*configType, error) {
	var config configType

	if _, err := toml.NewDecoder(reader).Decode(&config); err != nil {
		return nil, fmt.Errorf("unable to read package config: %w", err)
	}

	// Convert single-schema name to slice, for backward compatibility
	if config.Schema != "" && config.Schemas == nil {
		config.Schemas = []string{config.Schema}
	}

	for _, schemaName := range config.Schemas {
		if !schemaPattern.MatchString(schemaName) {
			return nil, fmt.Errorf("illegal schema name in pgpkg.toml: %s", schemaName)
		}
	}

	if err := CheckPackageName(config.Package); err != nil {
		return nil, err
	}

	for _, uses := range config.Uses {
		if err := CheckPackageName(uses); err != nil {
			return nil, err
		}
	}

	return &config, nil
}

func (c *configType) writeConfig(w io.Writer) error {
	// "schema" is deprecated, so don't export it, even if it's set.
	// To implement this, we make a copy of the config and then update it.
	config := *c
	config.Schema = ""

	encoder := toml.NewEncoder(w)
	return encoder.Encode(config)
}
