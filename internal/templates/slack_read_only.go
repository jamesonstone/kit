package templates

const slackReadOnlyGate = `## Slack: Read-Only by Default, Explicit Approval Required to Send

- Treat all Slack access as **read-only by default**.
- You may read and search Slack without additional approval. If given a Slack message or thread link, read the **entire thread**; inspect or search the containing channel and other relevant Slack content when useful.
- Use Slack content as reference material. Do **not** post, reply, react, edit, delete, forward, or otherwise modify Slack unless the human explicitly authorizes that specific action.
- Drafting a Slack message is not authorization to send it. Requests to draft, write, improve, or suggest a reply mean **draft only**; return the proposed text.
- Before any Slack send or other Slack mutation: draft the exact message or action, show the complete final content, ask whether the human authorizes **that specific action**, and wait for an explicit instruction such as **"send it," "send this," or "yes, send that message."**
- Approval is **single-use and message-specific**. Previous Slack send authorization and general statements such as "handle this" or "go ahead" do not authorize a later send unless they clearly refer to the exact message just presented.
- When uncertain whether the human authorized a Slack write, **do not perform it. Ask.**
- Load ` + "`docs/references/rules/slack-read-only.md`" + ` before any Slack write or when Slack investigation needs the full protocol.

`
