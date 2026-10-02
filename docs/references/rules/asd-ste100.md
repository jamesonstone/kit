---
kind: ruleset
slug: asd-ste100
description: Applies ASD-STE100 to agent-written explanatory prose while preserving technical meaning and explicit user style requests.
status: active
registry_scope: downstream
applies_to:
  - coding-agent
  - communication
  - planning
  - review
  - reporting
read_policy_default: must
---

# Ruleset: asd-ste100

## Purpose

Use ASD-STE100 Simplified Technical English for clear, consistent English explanations. Preserve technical precision and the detail that the user requests.

## Applies When

Apply by default to agent-authored responses, plans, progress updates, review findings, and completion reports. Read this rule before the first such response in a session. Read it again if the rule changes. This rule also applies when other Kit rules require explanations.

## Rules

### Writing and vocabulary

- Apply the writing rules and controlled dictionary of ASD-STE100 Issue 9. Use approved words with their approved meanings and parts of speech. Familiar plain English alone is not proof of ASD-STE100 compliance.
- Write short, direct sentences. Give one action per instruction. Put a condition before the action that it controls. Use the active voice for instructions. In explanations, use the active voice when the actor is known; do not invent an actor to remove a passive sentence.
- Use one technical term for one concept. Retain necessary software terminology under the standard's technical-noun and technical-verb rules. Use established project or subject-field terms. Do not classify an unfamiliar general word as a technical term merely to avoid dictionary restrictions.
- Preserve meaning, scope, uncertainty, and evidence. Do not omit a qualification or required detail to shorten a sentence. Separate verified facts, inferences, recommendations, and unresolved questions when they occur; short labels or separate sentences are sufficient. Keep a possible cause possible. Keep an unrun test unrun.
- If the official dictionary is available, review general words for their approved meaning and part of speech. If it is unavailable, do not invent approval or substitute a word that weakens technical meaning. Apply the verified official guidance and record the vocabulary-verification limit when material to task validation.

### Protected content and overrides

- Do not rewrite code, commands, API names, identifiers, paths, URLs, quoted source material, or exact error messages to satisfy this rule. Preserve required machine-readable formats. Apply the writing rule to explanatory prose around that content.
- If the user explicitly requests another language, tone, or writing style, follow that request instead of this writing default. The override changes prose style only; facts, technical tokens, evidence, permissions, approvals, and task-execution requirements still apply.
- Do not introduce a new approval step, refuse an authorized task, or reduce the requested detail because of this writing rule.

### Verification limits

- Do not call a response ASD-STE100 compliant because it reads clearly or passes structural checks. Full compliance needs review against the official writing rules and dictionary, including meanings, parts of speech, and permitted technical terms.
- Distinguish automated preservation/structure checks, dictionary review, and semantic or human review. A fixture test does not establish that an agent will produce the same response across providers.
- Document missing verification in the relevant validation record or completion report when it matters. Do not add repetitive compliance disclaimers to routine responses. No external service or runtime download is required.

## Verified dictionary examples

The official STEMG FAQ confirms these distinctions:

- Use `check` as a noun, not a general verb: write “Do a check of the result.”
- Use `about` for concerned with, not for an approximate amount. Retain the uncertainty when you write an approximate amount.
- Use `fall` for downward movement due to gravity, not for a decrease in a count or value.
- Use `start` instead of begin, commence, or initiate, as described in ASD's public guidance.

These examples are a small verified subset. They are not a replacement dictionary and do not certify the surrounding prose.

## Official references

- [STEMG standard overview](https://www.asd-ste100.org/about_STE.html): current issue, writing rules, dictionary, and technical terminology.
- [STEMG FAQ](https://www.asd-ste100.org/STE_faq.html): official rule explanations and dictionary examples.
- [ASD basics](https://www.asd-europe.org/standards-specifications/simplified-technical-english/what-are-the-basics-of-simplified-technical-english/): core writing guidance.
- [Official standard request](https://www.asd-ste100.org/STE_downloads.html): complete Issue 9 standard and dictionary. Do not submit personal data without permission.
- [STEMG AI white paper](https://www.asd-ste100.org/assets/files/WhitePaper-ASD-STE100_and_AI.pdf): AI assistance and verification limits.

Kit ships this rule, not the official standard or full dictionary. The implementation was reviewed against official public explanations and dictionary extracts. Full Issue 9 vocabulary verification was unavailable; the complete standard requires a request form. Do not infer complete coverage from the examples or tests.
