package cli

import "github.com/jamesonstone/kit/internal/config"

func defaultInitConfig() *config.Config {
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.DefaultInstructionScaffoldVersion
	return cfg
}

func projectInitDeliveryPaths(cfg *config.Config) []string {
	paths := []string{
		config.ConfigFileName,
		gitignorePath,
		makefilePath,
		codeRabbitConfigPath,
		pullRequestTemplatePath,
		autoAssignWorkflowPath,
		readmePath,
		cfg.ConstitutionPath,
	}
	for _, version := range []int{
		config.InstructionScaffoldVersionVerbose,
		config.InstructionScaffoldVersionTOC,
		config.InstructionScaffoldVersionMemory,
	} {
		paths = append(
			paths,
			instructionArtifactPaths(cfg, instructionFileSelection{}, version, true)...,
		)
	}
	return appendContextWorkflowDeliveryPaths(paths, cfg)
}
