// Package plan provides types and functions for parsing Terraform plan JSON output.
package plan

import "encoding/json"

// Action represents a Terraform resource action type.
type Action string

// Action constants matching Terraform's plan JSON output.
const (
	ActionCreate Action = "create"
	ActionRead   Action = "read"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionNoOp   Action = "no-op"
)

// Plan represents the top-level structure of a Terraform plan JSON file.
type Plan struct {
	FormatVersion    string               `json:"format_version"`
	TerraformVersion string               `json:"terraform_version"`
	PlannedValues    *StateValues         `json:"planned_values,omitempty"`
	ResourceChanges  []*ResourceChange    `json:"resource_changes,omitempty"`
	Configuration    *Configuration       `json:"configuration,omitempty"`
	PriorState       *StateValues         `json:"prior_state,omitempty"`
	Variables        map[string]*Variable `json:"variables,omitempty"`
}

// Variable represents a Terraform variable value.
type Variable struct {
	Value interface{} `json:"value"`
}

// StateValues represents planned values in the plan.
type StateValues struct {
	Outputs    map[string]*Output `json:"outputs,omitempty"`
	RootModule *StateModule       `json:"root_module,omitempty"`
}

// Output represents a Terraform output value.
type Output struct {
	Sensitive bool        `json:"sensitive,omitempty"`
	Value     interface{} `json:"value,omitempty"`
	Type      interface{} `json:"type,omitempty"`
}

// StateModule represents a module in the state.
type StateModule struct {
	Address      string           `json:"address,omitempty"`
	Resources    []*StateResource `json:"resources,omitempty"`
	ChildModules []*StateModule   `json:"child_modules,omitempty"`
}

// StateResource represents a resource in the planned state.
type StateResource struct {
	Address         string                 `json:"address"`
	Mode            string                 `json:"mode"`
	Type            string                 `json:"type"`
	Name            string                 `json:"name"`
	Index           interface{}            `json:"index,omitempty"`
	ProviderName    string                 `json:"provider_name"`
	SchemaVersion   int                    `json:"schema_version,omitempty"`
	Values          map[string]interface{} `json:"values,omitempty"`
	SensitiveValues interface{}            `json:"sensitive_values,omitempty"`
	DependsOn       []string               `json:"depends_on,omitempty"`
}

// ResourceChange represents a change to a resource in the plan.
type ResourceChange struct {
	Address         string      `json:"address"`
	PreviousAddress string      `json:"previous_address,omitempty"`
	ModuleAddress   string      `json:"module_address,omitempty"`
	Mode            string      `json:"mode"`
	Type            string      `json:"type"`
	Name            string      `json:"name"`
	Index           interface{} `json:"index,omitempty"`
	ProviderName    string      `json:"provider_name"`
	Change          *Change     `json:"change,omitempty"`
	ActionReason    string      `json:"action_reason,omitempty"`
	Deposed         string      `json:"deposed,omitempty"`
}

// Change represents the before and after values for a resource change.
type Change struct {
	Actions         []Action               `json:"actions"`
	Before          map[string]interface{} `json:"before,omitempty"`
	After           map[string]interface{} `json:"after,omitempty"`
	AfterUnknown    map[string]interface{} `json:"after_unknown,omitempty"`
	BeforeSensitive interface{}            `json:"before_sensitive,omitempty"`
	AfterSensitive  interface{}            `json:"after_sensitive,omitempty"`
	ReplacePaths    []interface{}          `json:"replace_paths,omitempty"`
}

// Configuration represents the Terraform configuration.
type Configuration struct {
	ProviderConfig map[string]*ProviderConfig `json:"provider_config,omitempty"`
	RootModule     *ConfigModule              `json:"root_module,omitempty"`
}

// ProviderConfig represents a provider's configuration.
type ProviderConfig struct {
	Name              string                 `json:"name"`
	FullName          string                 `json:"full_name,omitempty"`
	VersionConstraint string                 `json:"version_constraint,omitempty"`
	Expressions       map[string]interface{} `json:"expressions,omitempty"`
}

// ConfigModule represents a module in the configuration.
type ConfigModule struct {
	Outputs     map[string]*ConfigOutput   `json:"outputs,omitempty"`
	Resources   []*ConfigResource          `json:"resources,omitempty"`
	ModuleCalls map[string]*ModuleCall     `json:"module_calls,omitempty"`
	Variables   map[string]*ConfigVariable `json:"variables,omitempty"`
}

// ConfigOutput represents an output in the configuration.
type ConfigOutput struct {
	Expression  interface{} `json:"expression,omitempty"`
	Description string      `json:"description,omitempty"`
	Sensitive   bool        `json:"sensitive,omitempty"`
	DependsOn   []string    `json:"depends_on,omitempty"`
}

// ConfigResource represents a resource in the configuration.
type ConfigResource struct {
	Address           string                 `json:"address"`
	Mode              string                 `json:"mode"`
	Type              string                 `json:"type"`
	Name              string                 `json:"name"`
	ProviderConfigKey string                 `json:"provider_config_key"`
	Expressions       map[string]interface{} `json:"expressions,omitempty"`
	SchemaVersion     int                    `json:"schema_version,omitempty"`
	CountExpression   interface{}            `json:"count_expression,omitempty"`
	ForEachExpression interface{}            `json:"for_each_expression,omitempty"`
	DependsOn         []string               `json:"depends_on,omitempty"`
}

// ModuleCall represents a module call in the configuration.
type ModuleCall struct {
	Source            string                 `json:"source"`
	VersionConstraint string                 `json:"version_constraint,omitempty"`
	Expressions       map[string]interface{} `json:"expressions,omitempty"`
	CountExpression   interface{}            `json:"count_expression,omitempty"`
	ForEachExpression interface{}            `json:"for_each_expression,omitempty"`
	DependsOn         []string               `json:"depends_on,omitempty"`
	Module            *ConfigModule          `json:"module,omitempty"`
}

// ConfigVariable represents a variable in the configuration.
type ConfigVariable struct {
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description,omitempty"`
	Sensitive   bool        `json:"sensitive,omitempty"`
	Type        interface{} `json:"type,omitempty"`
}

// UnmarshalJSON provides custom unmarshaling for Action to handle string values.
func (a *Action) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*a = Action(s)
	return nil
}
