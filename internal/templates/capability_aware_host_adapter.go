package templates

const capabilityAwareHostAdapter = `## Capability-Aware Host Adapter

Use ` + "`docs/references/rules/agent-team-orchestration.md`" + ` for the canonical topology and lifecycle. This adapter translates semantic profiles into controls confirmed by the active host; host and model examples are illustrative, not fixed requirements.

| Profile | Capability target |
| --- | --- |
| ` + "`architect`" + ` | strongest justified architecture and systems reasoning |
| ` + "`orchestrator`" + ` | decomposition, dependency, risk, and overlap analysis |
| ` + "`mapper`" + ` | efficient read-heavy discovery and codebase mapping |
| ` + "`specialist`" + ` | bounded implementation with established contracts |
| ` + "`precision`" + ` | high-risk or ambiguous work requiring deeper judgment |
| ` + "`verifier`" + ` | fresh independent review with enough strength to challenge the implementation |

### Capability Negotiation

- Before delegation, inspect the host's confirmed child-launch, per-child model and effort, same-agent follow-up, wait or join, live roster or capacity, and fresh-verifier controls. Record confirmed absence as unavailable and unexposed controls as ` + "`unknown`" + `.
- Let the host govern concurrency. When parallel capacity is unconfirmed, do not invent a numeric cap or launch a child to probe it; a confirmed child primitive may still be used serially.
- If the host exposes no child primitive, use one supervisor with logical lanes. Never report a role prompt, task list, handoff, or manually opened conversation as a child.
- Resolve profiles against the live roster. Fall back to an equal-or-stronger eligible configuration, then a narrower low-risk lane with stronger verification, then a runtime-selected configuration reported as unverified; otherwise report ` + "`BLOCKED`" + `. An unavailable exact user model or effort pin remains blocked until the user changes it.
- If capacity changes or a spawn fails, keep unadmitted lanes pending, preserve accepted work, recompute the ready frontier from confirmed capacity, and report the failure and degradation. Do not retry indefinitely or let children create descendants.
- Report actual and logical lanes separately, requested and effective profiles plus model and effort when exposed, confirmed or unknown parallelism, continuity or replacement, verifier independence, and every fallback.

### Illustrative Current Mappings

| Host | Example profile mapping |
| --- | --- |
| Codex | strongest justified live configuration for ` + "`architect`" + ` and ` + "`precision`" + `; balanced read-heavy configuration for ` + "`orchestrator`" + ` and ` + "`mapper`" + `; fast configuration for bounded ` + "`specialist`" + ` work; fresh strong configuration for ` + "`verifier`" + ` |
| Claude Code | Opus-class for architecture, precision, and strong verification; Sonnet-class for orchestration, mapping, and general implementation; Haiku-class only for narrow low-risk bounded work |
| GitHub Copilot | request semantic profiles conservatively and record the effective model or agent controls the host confirms; do not infer child continuation, parallelism, or per-child selection |
| Warp/Oz | use native parent-child orchestration, continuation, and per-child model controls only when exposed; parallelism and admission remain host-owned |
| Unknown or single-agent host | serialize logical lanes in the supervisor and perform a distinct self-review, reporting that no child or independent verifier was confirmed |

Current provider references: [Codex subagents](https://learn.chatgpt.com/docs/agent-configuration/subagents), [Claude Code subagents](https://code.claude.com/docs/en/sub-agents), [GitHub Copilot custom-agent invocation](https://docs.github.com/en/copilot/how-tos/copilot-cli/use-copilot-cli/invoke-custom-agents), [Warp orchestration](https://docs.warp.dev/platform/orchestration/), and [Warp project rules](https://docs.warp.dev/agents/capabilities/rules/). Treat these as evidence sources to re-check, not as pinned capability promises.

`

func codexCapabilityAwareHostBinding(title string) string {
	if title != "AGENTS" {
		return ""
	}

	return `## Conditional Codex Subagent Binding

- Apply this section only when the active coding host is Codex. Warp/Oz and every other host that reads ` + "`AGENTS.md`" + ` must skip it.
- Before delegating, inspect the live Codex roster with ` + "`list_agents`" + `. The root supervisor may use ` + "`spawn_agent`" + ` with host-exposed ` + "`model`" + ` and ` + "`reasoning_effort`" + ` controls, ` + "`followup_task`" + ` for same-agent continuation, and ` + "`wait_agent`" + ` for status and joining; children must not spawn descendants.
- Resolve profiles from the live roster rather than static model IDs or a presumed capacity. If a native control is unavailable or fails, follow the shared host-adapter fallback and report the requested and effective profile, model, effort, continuity, and degradation.

`
}
