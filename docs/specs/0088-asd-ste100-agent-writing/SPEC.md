---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0088"
  slug: "asd-ste100-agent-writing"
  dir: "0088-asd-ste100-agent-writing"
---
# SPEC

## PURPOSE

Use ASD-STE100 as the default writing rule for agent-authored explanatory prose in Kit (GH-243). Make plans, updates, review findings, and reports easier to read. Preserve technical meaning, required detail, uncertainty, and evidence.

## CONTEXT

- Baseline: `ff9e043b67665237aff48b0372ce7d3c8ccf7c65`. The primary checkout is clean. The default branch was freshly fetched. Git and GitHub identities are Jameson Stone / jamesonstone.
- The universal contract is the canonical source for AGENTS.md, CLAUDE.md, and Copilot instructions. The generator renders the same managed block into each entry file. Cursor shares AGENTS.md, as recorded in spec 0079; no separate Cursor instruction block is needed. No repository-local `.agents/skills` directory or SKILL.md file exists in this checkout.
- Rules ship in the binary. Rules with `registry_scope: downstream` install by default; reconcile preserves local custom content. No general prose rule exists. `readme-header-tagline` concerns layout and is optional.
- PR #242 (`GH-241`, head `11f4cc3e63d3e17ac1df632f23fc47bfa01abcbe`) adds optional context/evidence guidance. This feature concerns how prose is written, not which evidence is selected. Keep that lane unchanged. Both changes add index rows; they can coexist, but merge order can require resolving that small shared-file conflict.

### Official sources and access

Consulted on 2026-10-02:

- [STEMG overview](https://www.asd-ste100.org/about_STE.html): current Issue 9, January 2025; writing rules, controlled dictionary, and technical noun/verb categories.
- [STEMG FAQ](https://www.asd-ste100.org/STE_faq.html): approved meanings and parts of speech, conditions before actions, active/passive voice, technical terms, and limits of general-purpose STE use. Dictionary examples include `check` as a noun, `about` with the meaning concerned with, and `fall` for gravity, not decrease. These are official dictionary extracts, not a complete lexicon.
- [ASD basics](https://www.asd-europe.org/standards-specifications/simplified-technical-english/what-are-the-basics-of-simplified-technical-english/): one instruction per sentence, consistent vocabulary, and `start` instead of begin/commence/initiate.
- [STEMG downloads](https://www.asd-ste100.org/STE_downloads.html): full Issue 9 standard and dictionary require a request form with personal data. No form was submitted. No complete standard/dictionary was found in the available local documents. The web tool returned 403 for official pages; direct read-only HTTP retrieval succeeded.
- [STEMG AI white paper](https://www.asd-ste100.org/assets/files/WhitePaper-ASD-STE100_and_AI.pdf), June 2026: apparent clarity is not proof of compliance; AI support does not replace informed review.

The full Issue 9 rule text and dictionary were NOT acquired. Investigation uses official public rule explanations and dictionary extracts. Full vocabulary verification is unavailable; do not call the examples or output fully compliant. Do not vendor the standard, fabricate dictionary entries, or add a runtime download/service requirement.

## REQUIREMENTS

- Default rule applies to English explanatory prose in responses, plans, progress updates, review findings, and completion reports across supported hosts.
- Use ASD-STE100 writing rules and approved vocabulary with approved meanings/parts of speech. Use short sentences and appropriate active voice. Give one action per instruction. Put controlling conditions first. Keep technical terms consistent.
- Retain needed software technical nouns/verbs under the standard's rules. Do not declare every unfamiliar word a technical term. Preserve meaning, scope, detail, uncertainty, and evidence.
- Distinguish facts, inferences, recommendations, and unresolved questions without forcing a report template on every response.
- Preserve code, commands, names, identifiers, paths, URLs, quotations, exact errors, and required machine-readable formats.
- Explicit user language, tone, and style requests override this writing default only. Preserve all permissions, approval requirements, and task behavior.
- State vocabulary-verification limits once in project validation/reporting when material. Do not add repeated disclaimers to routine responses. Structural checks must not imply full compliance.

## ACCEPTED PLAN

1. Add one downstream `asd-ste100` ruleset with official references, verified distinctions, technical-term handling, boundaries, and verification limits.
2. Add one short trigger/reference to the universal contract; render the existing provider entries from it. Use existing reconcile to register the rule in Kit's own managed state. Add one index row without importing PR #242's changes.
3. Add original before/after fixtures for five response types, plus explicit style/language overrides and structured-format/quotation cases. Add Go tests for default distribution and protected tokens, conditions, evidence/uncertainty anchors, and clear limits on automated assertions.
4. Review each fixture's meaning and uncertainty. Record this as agent semantic review, not human approval or full vocabulary review. Run required checks and deliver a ready human-assigned PR. Do not merge/deploy.

## DECISIONS

| Integration | Default reach | Canonical definition | Entry cost | Decision |
| --- | --- | --- | --- | --- |
| Full prose rules inline in universal contract | All rendered entries | Shared source, repeated full content | Large | Reject: crowds task context |
| Provider-specific instructions | Only edited providers | Duplicated | Variable | Reject: drift and inconsistent defaults |
| Optional ruleset | Opt-in projects | One rule | Small | Reject: does not meet default requirement |
| Downstream rule plus universal reference | All existing rendered entries | One rule | One short pointer | Select: reuses distribution and derivation |

- Use the existing source/template and rule registry, with no service, vocabulary database, runtime checker, or new configuration schema.
- Apply the standard as the default writing policy; separate that policy from evidence that a particular response complies. Public official extracts support limited vocabulary checks; full lexical/grammatical review remains unavailable.
- Respect the official FAQ's scope distinction: STE was designed for technical documentation. Applying it to conversational explanatory prose is Kit's chosen extension, with explicit user style/language overrides.
- Do not change existing safety-contract text to make it read as STE. This change controls new explanatory prose, not retroactive rewriting of invariants or technical tokens.
- Expected benefit: clearer actions and fewer ambiguous status/uncertainty statements. This is a hypothesis; no model-performance, comprehension, latency, or token-saving result is measured.

## DISCOVERIES

- Full standard access requires external data submission; public dictionary extracts are accessible without submission. Preserve that verification limit.
- The baseline binary reports an unrelated README refresh. Do not apply broad reconcile; target only this feature's rule and provider entries.

## VALIDATION

Observed local checks on 2026-10-02:

- Focused distribution, fixture, contract, and provider-generator tests: PASS.
- `gofmt -l .`: PASS (no output).
- `make vet`: PASS.
- `make lint`: PASS (0 issues).
- `make test`: PASS. The environment opt-out was removed for this command because usage-store tests exercise writes under temporary HOME; test binaries suppress ordinary telemetry.
- `make build`: PASS.
- `KIT_USAGE_DISABLED=1 ./bin/kit check --project`: PASS, with one nonblocking baseline compatibility advisory for project-owned content in the retired references index.
- `git diff --check`: PASS.
- PR #242 remains OPEN at its original head `11f4cc3e63d3e17ac1df632f23fc47bfa01abcbe`; no changes were made to that lane.

Fixtures: `pkg/cli/testdata/ste100/responses.json` contains eight before/after cases. Each includes an agent semantic-review note. The five required response types retain scope and uncertainty. Additional cases preserve fenced Go/JSON/YAML, exact error text, quotations, commands, identifiers, paths, and URLs. Spanish and conversational-style requests demonstrate explicit overrides. Tests assert protected content/counts, selected condition ordering, and evidence/uncertainty anchors. They evaluate curated examples and policy text, not live provider behavior or a rewriting engine.

Agent semantic review: COMPLETE for each fixture, with the limits above. Full dictionary/POS checking, human language review, and live cross-provider response tests: UNRUN. Performance/readability benefit: UNMEASURED hypothesis. No test establishes full ASD-STE100 compliance.

## OUTCOME

Implementation and local validation complete. Issue #243; branch/worktree GH-243. Ready pull-request delivery follows validation. Merge and deployment are outside scope.

## REPOSITORY MEMORY

Adopted this living spec before implementation. Durable writing guidance belongs in the canonical ruleset. No new project-wide invariant needs a Constitution edit.
