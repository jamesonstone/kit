package cli

import "github.com/jamesonstone/kit/v3/internal/config"

func defaultInitConfig() *config.Config {
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.CurrentInstructionScaffoldVersion
	return cfg
}
