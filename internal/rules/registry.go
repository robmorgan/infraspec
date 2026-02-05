package rules

// Registry holds all registered rules.
type Registry struct {
	rules []Rule
}

// NewRegistry creates a new empty rule registry.
func NewRegistry() *Registry {
	return &Registry{
		rules: make([]Rule, 0),
	}
}

// Register adds a single rule to the registry.
func (r *Registry) Register(rule Rule) {
	r.rules = append(r.rules, rule)
}

// RegisterAll adds multiple rules to the registry.
func (r *Registry) RegisterAll(rules ...Rule) {
	r.rules = append(r.rules, rules...)
}

// RulesForResource returns all rules that apply to the given resource type.
func (r *Registry) RulesForResource(resourceType string) []Rule {
	var matched []Rule
	for _, rule := range r.rules {
		if rule.ResourceType() == resourceType {
			matched = append(matched, rule)
		}
	}
	return matched
}

// RulesForProvider returns all rules that apply to the given provider.
func (r *Registry) RulesForProvider(provider string) []Rule {
	var matched []Rule
	for _, rule := range r.rules {
		if rule.Provider() == provider {
			matched = append(matched, rule)
		}
	}
	return matched
}

// AllRules returns all registered rules.
func (r *Registry) AllRules() []Rule {
	return r.rules
}

// RuleByID looks up a rule by its unique identifier.
// Returns the rule and true if found, nil and false otherwise.
func (r *Registry) RuleByID(id string) (Rule, bool) {
	for _, rule := range r.rules {
		if rule.ID() == id {
			return rule, true
		}
	}
	return nil, false
}

// DefaultRegistry returns a registry pre-populated with built-in rules.
// Currently returns an empty registry as no built-in rules are implemented yet.
func DefaultRegistry() *Registry {
	return NewRegistry()
}
