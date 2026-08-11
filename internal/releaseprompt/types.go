package releaseprompt

type ValueSource string

const (
	SourceExplicit   ValueSource = "explicit"
	SourceDiscovered ValueSource = "discovered"
	SourceDefaulted  ValueSource = "defaulted"
)

type Input struct {
	Repositories           []string
	Root                   string
	Project                string
	Organization           string
	FeatureContext         string
	ScopeExpansion         string
	InfrastructureMode     string
	InfrastructureProvider string
	InfrastructureCLI      string
	Environment            string
	ProductionVerification string
	IntegrationSuite       string
}

type Repository struct {
	Path                   string `yaml:"path"`
	Name                   string `yaml:"name"`
	GitHub                 string `yaml:"github,omitempty"`
	DefaultBranch          string `yaml:"default_branch,omitempty"`
	KitManaged             bool   `yaml:"kit_managed"`
	ReleaseWorkflowPresent bool   `yaml:"release_workflow_present"`
	VerificationHint       string `yaml:"verification_hint,omitempty"`
	IntegrationSuiteHint   string `yaml:"integration_suite_hint,omitempty"`
}

type SourceControl struct {
	Provider string `yaml:"provider"`
	CLI      string `yaml:"cli"`
}

type Infrastructure struct {
	Mode          string `yaml:"mode"`
	Provider      string `yaml:"provider"`
	CLI           string `yaml:"cli"`
	IdentityCheck string `yaml:"identity_check"`
	Policy        string `yaml:"policy"`
}

type Production struct {
	Environment  string `yaml:"environment"`
	Verification string `yaml:"verification"`
}

type FieldResolution struct {
	Field  string      `yaml:"field"`
	Value  string      `yaml:"value"`
	Source ValueSource `yaml:"source"`
}

type Config struct {
	Project                 string            `yaml:"project"`
	Repositories            []Repository      `yaml:"repositories"`
	ScopeExpansion          string            `yaml:"scope_expansion"`
	Organization            string            `yaml:"organization"`
	FeatureContext          string            `yaml:"feature_context"`
	SourceControl           SourceControl     `yaml:"source_control"`
	Infrastructure          Infrastructure    `yaml:"infrastructure"`
	Production              Production        `yaml:"production"`
	IntegrationSuite        string            `yaml:"integration_suite"`
	DeploymentContext       string            `yaml:"deployment_context"`
	ReviewSystems           string            `yaml:"review_systems"`
	RequiredChecks          string            `yaml:"required_checks"`
	DatabaseMigrationPolicy string            `yaml:"database_migration_policy"`
	AdditionalHardRules     string            `yaml:"additional_hard_rules"`
	FinalReportRequirements string            `yaml:"final_report_requirements"`
	Resolution              []FieldResolution `yaml:"resolution"`
}

func (c *Config) record(field, value string, source ValueSource) {
	c.Resolution = append(c.Resolution, FieldResolution{Field: field, Value: value, Source: source})
}
