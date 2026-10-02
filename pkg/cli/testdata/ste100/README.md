# ASD-STE100 response fixtures

These original examples illustrate Kit's writing policy. They are not certified ASD-STE100 text. The complete Issue 9 standard and dictionary were unavailable during implementation.

`responses.json` has before/after examples for a coding plan, progress update, review finding, blocked task, and completion report. It also has examples for protected source content, required structured output, and explicit Spanish and conversational-style requests.

Each `review` field records an agent review of meaning, scope, evidence, and uncertainty. The review is a reasoned assessment, not an automated proof or human review. Technical content can appear in a different position when a controlling condition moves before an action. Its content and occurrence count must remain unchanged.

Automated tests cover:

- Distribution through fresh init and existing-project reconcile, including provider references and preservation of local rule changes.
- Exact preservation of fenced code/JSON/YAML, inline technical tokens, and quotations.
- Evidence and uncertainty anchors, ordered plan steps, and condition placement in selected examples.
- Presence of explicit override requests and their expected Spanish or conversational expressions.

The tests evaluate curated examples and policy text. They do not invoke agents, implement a rewrite engine, or prove that every provider follows the rule. Structural checks do not establish approved word meanings, parts of speech, permitted technical terms, or semantic equivalence. Full dictionary review, human review, and live cross-provider validation remain unrun.

Example: “The focused tests passed, but hosted CI is `PENDING` and integration tests were not run” becomes “The focused tests passed. Hosted CI is `PENDING`. Integration tests were not run.” The three result states remain distinct.
