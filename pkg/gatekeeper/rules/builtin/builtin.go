// Package builtin provides embedded built-in security rules.
package builtin

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/robmorgan/infraspec/pkg/gatekeeper/rules"
)

//go:embed rules/*.yaml
var builtinRulesFS embed.FS

// LoadBuiltinRules loads all built-in rules from embedded files
func LoadBuiltinRules() ([]rules.Rule, error) {
	var allRules []rules.Rule

	err := fs.WalkDir(builtinRulesFS, "rules", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		data, err := builtinRulesFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		fileRules, err := rules.LoadFromBytes(data)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		allRules = append(allRules, fileRules...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return allRules, nil
}
