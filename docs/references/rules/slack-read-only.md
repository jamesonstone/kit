---
kind: ruleset
slug: slack-read-only
description: 'Treat Slack as read-only by default and require explicit, message-specific human approval before any Slack send or other Slack mutation.'
status: active
registry_scope: downstream
applies_to:
  - slack
  - messaging
  - communication
  - coding-agent
  - automation
  - collaboration
read_policy_default: must
---

# Ruleset: Slack Read-Only

## Purpose

- Treat all Slack access as read-only by default.
- Allow reading, searching, and using Slack content as reference material
  without additional approval.
- Prohibit posting, replying, reacting, editing, deleting, forwarding, or
  otherwise modifying Slack unless the human explicitly authorizes that
  specific action.

## Applies When

Load this ruleset immediately when:

- The human provides a Slack link, mentions a Slack channel, thread, or
  message, or asks to search or investigate Slack.
- The agent considers posting, replying, reacting, editing, deleting,
  forwarding, sharing, or otherwise modifying Slack state.
- Slack tools, MCP servers, CLIs, or APIs are available for the current
  session.

Reading, searching, retrieving context, and analyzing Slack content do not
require approval and remain read-only.

## Rules

### Read-Only By Default

- You may read and search Slack without additional approval.
- If the human provides a direct link to a Slack message or thread, read the
  entire thread for context.
- Also inspect or search the channel containing that thread when additional
  context may be relevant.
- You may search other relevant Slack channels, threads, and messages when
  useful for understanding the request.
- Use Slack content as reference material for analysis, investigation,
  summaries, debugging, or drafting.

### Never Send Without Explicit Approval

Never send a Slack message without the human's explicit approval.

Drafting a Slack message is not authorization to send it.

Requests such as "draft a response," "write a reply," "what should I say?",
"help me respond," or "make this message better" mean draft only. Return the
proposed text. Do not send it.

If you believe a Slack message should be sent:

1. Draft the exact message.
2. Show the complete final message before sending.
3. Ask whether the human authorizes sending that specific message.
4. Wait for an explicit instruction such as "send it," "send this," or
   "yes, send that message."
5. Only then may you perform the Slack send action.

Approval is single-use and message-specific. Previous authorization to send
another Slack message does not authorize future messages. General statements
such as "handle this," "take care of it," "respond to this," or "go ahead"
must not be treated as authorization to send unless they clearly refer to the
exact Slack message just presented for approval.

When uncertain whether the human authorized a Slack write action, do not
perform it. Ask.

### Other Slack Mutations

The same explicit-approval requirement applies to:

- posting a new message
- replying in a thread
- editing a message
- deleting a message
- adding or removing reactions
- forwarding or sharing a message
- changing channel information
- performing any other action that modifies Slack state

## Anti-Patterns

- Treating a draft, rewrite, or "what should I say?" request as send
  authorization.
- Reusing prior Slack send approval for a later message.
- Interpreting "handle this," "take care of it," "respond to this," or
  "go ahead" as send authorization unless they clearly refer to the exact
  message just presented.
- Posting, reacting, editing, or deleting Slack content merely because a
  Slack tool is available.
- Inlining this entire ruleset into always-loaded instruction files.

## Verification

Before any Slack write:

- Confirm the complete final message or mutation was shown to the human.
- Confirm the human explicitly authorized that specific action with language
  such as "send it," "send this," or "yes, send that message."
- Confirm the authorization is for this message or mutation, not a previous
  one.
- If any of those are missing, do not write to Slack. Ask.

## Examples

Allowed without approval:

```text
Read the linked Slack thread and search the containing channel for related
context.
```

Draft only:

```text
Here is a proposed reply. I have not sent it. Say "send this" if you want me
to post it.
```

Forbidden:

```text
I drafted a reply and posted it to the thread.
```
