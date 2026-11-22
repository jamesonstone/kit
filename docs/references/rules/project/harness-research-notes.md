---
kind: ruleset
slug: harness-research-notes
description: Pairs Kit changes with concise research posts on jamesonstone.io.
status: active
applies_to:
  - coding-agent
  - pull-request
  - publication
read_policy_default: conditional
---

# Ruleset: harness-research-notes

## Purpose

Record what changed in Kit, why it changed, and what the evidence supports in the public series **Coding-Agent Harness Research**.

## Applies When

- Before opening or merging a Kit pull request for a feature, fix, refactor, tweak, or design change.
- After the Kit change merges, until its companion post is published and verified.

## Rules

- Draft a short research post for each change. A single post can cover a related set if it names every change and its reason. Prefix the article title with **Coding-Agent Harness Research** and use the next series number.
- Start with the question or observation that prompted the work. Summarize the shipped changes in plain language, explain the decision and tradeoff, then state what the checks showed and what remains untested. Prefer one concrete example. Link public source PRs, releases, or primary evidence. Present hypotheses as hypotheses. Use only information suitable for public release.
- Add the post under `_posts/` in [jamesonstone/jamesonstone.github.io](https://github.com/jamesonstone/jamesonstone.github.io) and follow that repository's delivery, editorial, and validation rules. Keep the Kit PR link and website PR link together in the delivery record. Open the website PR alongside the Kit change and reconcile its claims against the merged Kit source before site publication.
- After Kit merges, publish through the site's existing GitHub Pages workflow under the applicable site merge authorization. If publication is pending or fails, keep the site PR or recovery action linked and report the actual state.
- This rule belongs to Kit's source repository. Keep it outside the binary's embedded rule files, the universal contract, and downstream scaffolds.

## Verification

- Confirm that the site build renders the intended permalink and that the post links the Kit change and describes the observed result.
- After the site publishes, open the public article on `jamesonstone.io` and record the verified URL in the Kit delivery record. A website source commit or a successful build is a separate milestone.
