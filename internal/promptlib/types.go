package promptlib

type SourceKind string

const (
	SourceBuiltin SourceKind = "builtin"
	SourceGlobal  SourceKind = "global"
	SourceLocal   SourceKind = "local"
)

type Identity struct {
	Noun string
	Verb string
}

func (i Identity) CommandName() string {
	if i.Noun == "" {
		return i.Verb
	}
	if i.Verb == "" {
		return i.Noun
	}
	return i.Noun + " " + i.Verb
}

type RenderFunc func() (string, error)

type Prompt struct {
	Identity            Identity
	Content             string
	Description         string
	ContextRequirements []string
	Render              RenderFunc
}

type Source struct {
	Kind     SourceKind
	Location string
	Prompts  []Prompt
}

type SourcePrompt struct {
	Prompt   Prompt
	Kind     SourceKind
	Location string
}

type EffectivePrompt struct {
	Prompt   Prompt
	Kind     SourceKind
	Location string
	Shadowed []SourcePrompt
}

func (p EffectivePrompt) CommandName() string {
	return p.Prompt.Identity.CommandName()
}

func (p EffectivePrompt) ShadowSummary() string {
	if len(p.Shadowed) == 0 {
		return ""
	}

	kinds := make([]string, 0, len(p.Shadowed))
	for _, shadowed := range p.Shadowed {
		kinds = append(kinds, string(shadowed.Kind))
	}
	return string(p.Kind) + " overrides " + joinKinds(kinds)
}
