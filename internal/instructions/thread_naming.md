# Conversation naming

A conversation title identifies the work it currently owns, not its first
prompt. Naming changes metadata only; preserve the same conversation ID,
history, context/cache and tool state.

## Canonical policy

Use `[scope] domain / objective`. Aim for 60 characters or fewer when practical;
keep meaning rather than truncating it. Preserve useful acronyms such as EUID.

- **Scope:** reuse the owning repository from supplied project context. For
  genuine cross-repository work, reuse an existing program/workspace name or
  the smallest stable conceptual scope. Never concatenate repository lists.
  Do not derive scope from an incidental shell directory, worktree lane or
  issue number. If ownership is unclear, keep the title until it is known.
- **Domain:** a stable conceptual noun or compact noun phrase, such as eventing,
  receiving, accessioning, identity or temporal. Keep it through stage changes.
- **Objective:** the current desired outcome or unresolved problem. Avoid vague
  labels such as fixes, cleanup, investigation, testing, follow-up, task 2 or
  misc unless they actually identify the owned outcome.

## Checkpoints and decision

Evaluate on creation, on a user-requested fork once its child objective is
known, and at an already-occurring planning/handoff/major-outcome checkpoint
where ownership may have changed. On resume or after host-owned compaction,
use the context already provided; do not load a transcript for naming.
Do not schedule polling, add model calls, or run a hook after every turn.

Apply this ordered decision to the exact current conversation:

1. Honor an explicit user instruction to keep a title or stop renaming.
   Sidebar pinning does not lock a title. Preserve pin state.
2. If the current title still describes the owned problem/outcome, do nothing.
   Editing, tests, deployment and log inspection alone do not change ownership.
3. If evidence of a material change is insufficient, keep the meaningful title.
4. Otherwise derive scope, domain and objective from available context. Compare
   against the current title ignoring case, repeated whitespace and cosmetic
   punctuation. Do nothing if effectively unchanged. Preserve meaningful token
   boundaries; different numbers, identities and objectives remain different.
5. If changed, use the active host adapter below. Never invoke a rename with an
   effectively unchanged title. Remember the last applied/suggested title in
   the existing conversation context, not in a new persistent store.

The active agent interprets meaning. The optional read-only helper
`kit instructions title --scope labcore --domain eventing --objective 'prod delivery'`
formats already-resolved fields; it does not infer meaning from arbitrary work
text. Pass `--current` and `--accurate` when the current title still owns the
outcome. `--json` exposes the decision; normal output is empty when unchanged.
It is not necessary to call this helper at every checkpoint.

## Forks

Rename only the child, using its verified identity and the objective that
caused the fork. Keep the parent unchanged. Do not append 2, fork or branch,
encode ancestry, or create another conversation for naming. If the fork has
not returned a stable child ID, wait for the child to initialize itself.
A resumed conversation is not automatically a fork. Retain useful inherited
context; no lineage database is needed.

## Host capability adapters

Use the first safe supported mechanism actually exposed by the current host.
A product name, MCP support, installed CLI or manual slash command does not
prove an agent-callable rename API. Unknown capability routes as unavailable.

- **A — automatic:** when a supported tool/API is available and bound to the
  exact current/authorized child identity, apply the changed title once and
  verify returned metadata or supported readback. No new permission prompt is
  required for this metadata-only operation within the naming request.
- **B — assisted:** naming exists but automatic application cannot be safely
  established. Emit one title and the exact supported manual action.
- **C — policy only:** no safe mechanism is established. Emit one suggested
  title at a materially stale transition or explicit request, then continue.

On a failed rename, retain conversation state and report the failure once with
an assisted fallback. Do not retry until capability/identity evidence changes.
If current title/readback is unavailable, use the last confirmed title in the
current context. If none is available, suggest once; do not repeatedly guess.
Creation with an absent title and a known objective may be named once.

### Codex Mac and CLI

When exposed, use `set_thread_title` (or the host's equivalent native setter).
Prefer its implicit calling-task identity; provide an explicit child ID only
when returned by the authorized fork. Verify the result. A host already
exposing a supported app-server connection may use `thread/name/set` with
`threadId` and `name`. Do not start another server or resume a session solely
to rename it. Without that binding, offer `/rename <title>` in the CLI.

Codex's separate initialization rule still orders initial rename before pin;
it does not forbid later semantic title updates. If the initial prompt is an
unread attachment, defer semantic derivation until it has been read rather
than inventing its objective. Preserve the existing pin behavior.

### Claude Code CLI and Code in Desktop

Use an exposed native rename tool if available. Otherwise, an already-available
supported Agent SDK can use `renameSession(sessionId, title, {dir})`, with
`getSessionInfo(sessionId, {dir})` readback. Bind the ID to the current session
from host-provided identity (for example documented `${CLAUDE_SESSION_ID}` in
skill content) or a supported lifecycle event's `session_id`. Never choose the
most recent session, enumerate unrelated sessions, or scrape private files.
Use the original session directory, not an incidental worktree cwd. If SDK,
identity or readback cannot be established safely, offer `/rename <title>`.
Do not launch a nested Claude session or install dependencies merely to rename.

An existing SessionStart integration with a known semantic title may return
`hookSpecificOutput: {hookEventName: "SessionStart", sessionTitle: title}` on
startup/resume/fork. Preserve `session_title` when accurate; omit a title when
new ownership is unknown. This field is ignored for clear/compact. Kit installs
no startup inference hook: the user's next objective may not yet be available.
Desktop Code shares hooks/settings; verify sidebar reflection when available.
Ordinary Claude Chat/Cowork is a separate host: do not assume it loads CLAUDE.md
or exposes Code's SDK/hooks. Use the suggested-title fallback when unsupported.

### Cursor Mac, CLI and Cloud Agents

Cursor reads repository AGENTS.md; no duplicate semantic policy is needed.
Use a native rename tool only if the active host actually exposes one with a
verified current identity. Otherwise, in the editor instruct the user to open
chat history and click the title to rename; in CLI offer `/rename <title>`.
Cursor's sessionStart hook has no verified title output; do not invent one.
For an independently authorized Cloud Agent creation, its documented `name`
field can carry the canonical title. That creation API does not establish an
existing-agent rename API; never create a replacement to achieve a new title.

## Manual fallback

At a materially stale transition or explicit request, emit exactly one title:

```text
Suggested thread title:
[labcore] eventing / prod delivery verification
```

Add `/rename <title>` or the supported manual UI action when applicable. Do not
repeat a suggestion during ordinary work or claim it was applied.

## Safety boundary

Never manipulate undocumented SQLite/databases, private host state, Electron
internals, accessibility/macOS UI automation or host conversation files for
naming. Never restart, replace, duplicate, unnecessarily fork, compact,
summarize or discard tool state for naming. Rename must not alter repository
state. Supported SDK/API metadata operations are allowed; reproducing their
private storage implementation is not. Context preservation does not promise
that a provider retains a particular billing/cache hit across ordinary turns.

## Capability evidence (2026-09-15)

VERIFIED means observed here or explicitly documented; SUPPORTED BUT LIMITED
means a documented mechanism has runtime preconditions. UNKNOWN is not support;
UNSUPPORTED applies to a known absent mechanism, not to an entire product.

| Host | Policy installation | Agent can rename | Mechanism | Fork/resume | Limitations |
| --- | --- | --- | --- | --- | --- |
| Codex Mac | VERIFIED AGENTS.md | VERIFIED in this task | native set_thread_title | exact child ID; same resumed ID | tools vary by host |
| Codex CLI | VERIFIED AGENTS.md | SUPPORTED BUT LIMITED | bound app-server API; manual /rename | documented resume/fork | CLI presence alone is not API binding |
| Claude Code | VERIFIED CLAUDE.md | SUPPORTED BUT LIMITED | SDK renameSession; SessionStart.sessionTitle | startup/resume/fork hook; SDK exact ID | SDK and identity required; hook needs known title |
| Claude Code Desktop | VERIFIED CLAUDE.md/shared hooks | SUPPORTED BUT LIMITED | Code hooks; toolbar rename | verify actual session identity | automatic sidebar reflection UNOBSERVED |
| Claude Chat/Cowork | UNKNOWN repo loading | UNKNOWN | suggested title | UNKNOWN | do not equate with Code |
| Cursor Mac Agent | VERIFIED AGENTS.md | UNKNOWN native API | assisted history-title rename | child uses its own objective | manual UI is not agent automation |
| Cursor CLI | VERIFIED AGENTS.md | UNKNOWN native API | assisted /rename | documented /resume and /fork | no verified title hook |
| Cursor Cloud | repository rules per host | creation only | name on authorized creation | existing rename UNKNOWN | never recreate to rename |

Sources: [Codex app-server](https://learn.chatgpt.com/docs/app-server),
[Codex CLI](https://learn.chatgpt.com/docs/developer-commands?surface=cli),
[Claude sessions](https://code.claude.com/docs/en/sessions),
[Claude SDK](https://code.claude.com/docs/en/agent-sdk/typescript#renamesession),
[Claude hooks](https://code.claude.com/docs/en/hooks#sessionstart),
[Claude skills identity](https://code.claude.com/docs/en/skills),
[Claude Desktop Code](https://code.claude.com/docs/en/desktop),
[Cursor rules](https://prod.cursor.com/docs/rules),
[Cursor history](https://prod.cursor.com/docs/agent/chat/history),
[Cursor CLI](https://prod.cursor.com/docs/cli/reference/slash-commands),
[Cursor hooks](https://prod.cursor.com/docs/hooks),
[Cursor Cloud API](https://prod.cursor.com/docs/cloud-agent/api/endpoints).
Recheck exposed capabilities when changing hosts; do not rediscover them each turn.
