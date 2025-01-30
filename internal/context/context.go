package context

import (
	"github.com/gruntwork-io/terratest/modules/terraform"

	"github.com/robmorgan/infraspec/internal/config"
	"github.com/robmorgan/infraspec/pkg/assertions"
)

type TestContext struct {
	config       *config.Config
	tfOptions    *terraform.Options
	storedValues map[string]string
	assertions   map[string]assertions.Asserter
	cleanup      []func() error
	scenarioUri  string
}

func New(cfg *config.Config) *TestContext {
	return &TestContext{
		config:       cfg,
		storedValues: make(map[string]string),
		assertions:   make(map[string]assertions.Asserter),
		cleanup:      make([]func() error, 0),
	}
}

func (t *TestContext) Config() *config.Config {
	return t.config
}

func (t *TestContext) GetTerraformOptions() *terraform.Options {
	return t.tfOptions
}

func (t *TestContext) SetTerraformOptions(opts *terraform.Options) {
	t.tfOptions = opts
}

func (t *TestContext) SetScenarioUri(uri string) {
	t.scenarioUri = uri
}

func (t *TestContext) GetScenarioUri() string {
	return t.scenarioUri
}

func (t *TestContext) StoreValue(key, value string) {
	t.storedValues[key] = value
}

func (t *TestContext) GetStoredValues() map[string]string {
	return t.storedValues
}

func (t *TestContext) GetValue(key string) (string, bool) {
	value, exists := t.storedValues[key]
	return value, exists
}

func (t *TestContext) AddCleanup(fn func() error) {
	t.cleanup = append(t.cleanup, fn)
}

func (t *TestContext) Cleanup() error {
	// Execute cleanup functions in reverse order
	for i := len(t.cleanup) - 1; i >= 0; i-- {
		if err := t.cleanup[i](); err != nil {
			return err
		}
	}
	return nil
}

func (t *TestContext) GetAsserter(provider string) (assertions.Asserter, error) {
	if asserter, exists := t.assertions[provider]; exists {
		return asserter, nil
	}

	// Create new asserter based on provider
	asserter, err := assertions.New(provider, t.config.DefaultRegion)
	if err != nil {
		return nil, err
	}

	t.assertions[provider] = asserter
	return asserter, nil
}
