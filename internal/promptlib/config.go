package promptlib

import "github.com/jamesonstone/kit/internal/config"

func SourceFromConfig(kind SourceKind, location string, cfg *config.Config) Source {
	if cfg == nil || len(cfg.Prompts) == 0 {
		return Source{Kind: kind, Location: location}
	}

	var prompts []Prompt
	for noun, verbs := range cfg.Prompts {
		for verb, prompt := range verbs {
			prompts = append(prompts, Prompt{
				Identity: Identity{
					Noun: noun,
					Verb: verb,
				},
				Content:     prompt.Content,
				Description: prompt.Description,
			})
		}
	}

	return Source{
		Kind:     kind,
		Location: location,
		Prompts:  prompts,
	}
}
