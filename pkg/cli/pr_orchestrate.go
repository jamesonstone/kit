package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/releaseprompt"
)

type prOrchestrateOptions struct {
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
	DryRun                 bool
	OutputOnly             bool
	Copy                   bool
}

var (
	prOrchestrateRunner           releaseprompt.Runner = releaseprompt.SystemRunner{}
	prOrchestrateResolve                               = releaseprompt.Resolve
	prOrchestrateRender                                = releaseprompt.Render
	prOrchestrateRenderDryRun                          = releaseprompt.RenderDryRun
	prOrchestrateInteractiveCheck                      = streamsHaveInteractiveTerminal
)

func newPROrchestrateCommand() *cobra.Command {
	opts := prOrchestrateOptions{}
	cmd := &cobra.Command{
		Use:           "orchestrate",
		Short:         "Generate a dependency-aware release orchestration prompt",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		Long: `Generate a fully resolved coding-agent prompt for dependency-aware,
multi-repository pull-request release orchestration.

Kit performs bounded read-only repository discovery and renders the prompt; it
does not enumerate the release set, merge pull requests, deploy software,
mutate infrastructure, or launch an agent. Use repeatable --repos for exact
repositories or --root for the root itself plus immediate child repositories.
Without either flag, an interactive terminal asks for scope. Redirected input
requires explicit scope. Non-terminal stdout receives only the raw prompt.
Use --dry-run to inspect deterministic configuration provenance and the prompt
without clipboard access or release-side mutation.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runPROrchestrate(cmd, opts)
		},
	}
	flags := cmd.Flags()
	flags.StringArrayVar(&opts.Repositories, "repos", nil, "exact repository path; repeat for multiple repositories")
	flags.StringVar(&opts.Root, "root", "", "bounded root containing the root repository and immediate child repositories")
	flags.StringVar(&opts.Project, "project", "", "project or release program name; inferred when omitted")
	flags.StringVar(&opts.Organization, "organization", "", "source-control organization; inferred when omitted")
	flags.StringVar(&opts.FeatureContext, "context", "", "optional feature or release context; the agent infers it when omitted")
	flags.StringVar(&opts.ScopeExpansion, "scope", "", "scope expansion policy: strict, related, or organization (default related)")
	flags.StringVar(&opts.InfrastructureMode, "infra", "", "infrastructure mode: auto, none, direct, iac, mixed, or custom")
	flags.StringVar(&opts.InfrastructureProvider, "infra-provider", "", "infrastructure provider such as aws, azure, or gcp")
	flags.StringVar(&opts.InfrastructureCLI, "infra-cli", "", "infrastructure CLI or source-of-truth command")
	flags.StringVar(&opts.Environment, "environment", "", "single target environment (default production)")
	flags.StringVar(&opts.ProductionVerification, "verify", "", "auto, command:, script:, endpoint:, or instruction:")
	flags.StringVar(&opts.IntegrationSuite, "integration-suite", "", "auto, command:, script:, endpoint:, instruction:, or none")
	flags.BoolVar(&opts.DryRun, "dry-run", false, "show resolved configuration provenance and prompt without clipboard access")
	flags.BoolVar(&opts.OutputOnly, "output-only", false, "write raw prompt text to stdout instead of the interactive clipboard default")
	flags.BoolVar(&opts.Copy, "copy", false, "copy the prompt in addition to requested stdout output")
	return cmd
}

func runPROrchestrate(cmd *cobra.Command, opts prOrchestrateOptions) error {
	interactive := prOrchestrateInteractiveCheck(cmd.InOrStdin(), cmd.ErrOrStderr())
	if len(opts.Repositories) == 0 && strings.TrimSpace(opts.Root) == "" {
		if !interactive {
			return fmt.Errorf("repository scope is required in noninteractive mode; use --repos or --root")
		}
		repositories, root, err := promptPROrchestrateScope(cmd.InOrStdin(), cmd.ErrOrStderr())
		if err != nil {
			return err
		}
		opts.Repositories, opts.Root = repositories, root
	}
	config, err := prOrchestrateResolve(cmd.Context(), releasePromptInput(opts), prOrchestrateRunner)
	if err != nil {
		return err
	}
	prompt, err := prOrchestrateRender(config)
	if err != nil {
		return err
	}
	if opts.DryRun {
		bundle, err := prOrchestrateRenderDryRun(config, prompt)
		if err != nil {
			return err
		}
		_, err = io.WriteString(cmd.OutOrStdout(), bundle)
		return err
	}
	return writePROrchestratePrompt(cmd, prompt, interactive, opts.OutputOnly, opts.Copy)
}

func releasePromptInput(opts prOrchestrateOptions) releaseprompt.Input {
	return releaseprompt.Input{
		Repositories: opts.Repositories, Root: opts.Root, Project: opts.Project,
		Organization: opts.Organization, FeatureContext: opts.FeatureContext,
		ScopeExpansion: opts.ScopeExpansion, InfrastructureMode: opts.InfrastructureMode,
		InfrastructureProvider: opts.InfrastructureProvider, InfrastructureCLI: opts.InfrastructureCLI,
		Environment: opts.Environment, ProductionVerification: opts.ProductionVerification,
		IntegrationSuite: opts.IntegrationSuite,
	}
}

func promptPROrchestrateScope(in io.Reader, out io.Writer) ([]string, string, error) {
	if _, err := fmt.Fprint(out, "Repository scope [root:.] (use repos:path1,path2 for exact repositories): "); err != nil {
		return nil, "", err
	}
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && len(line) == 0 {
		return nil, "", fmt.Errorf("read repository scope: %w", err)
	}
	value := strings.TrimSpace(line)
	if value == "" {
		return nil, ".", nil
	}
	if strings.HasPrefix(value, "repos:") {
		var repositories []string
		for _, repository := range strings.Split(strings.TrimPrefix(value, "repos:"), ",") {
			if repository = strings.TrimSpace(repository); repository != "" {
				repositories = append(repositories, repository)
			}
		}
		if len(repositories) == 0 {
			return nil, "", fmt.Errorf("interactive repository list is empty")
		}
		return repositories, "", nil
	}
	return nil, strings.TrimPrefix(value, "root:"), nil
}
