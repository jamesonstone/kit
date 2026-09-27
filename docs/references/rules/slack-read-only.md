---
kind: ruleset
slug: slack-read-only
description: Treats Slack as read-only and requires explicit message-specific approval for every write.
status: active
registry_scope: downstream
applies_to:
  - slack
  - messaging
read_policy_default: conditional
---

# Ruleset: slack-read-only

## Purpose

- Never speak for the human in Slack without their explicit approval.

## Applies When

- Reading a Slack link, searching Slack, or any Slack write.

## Rules

- Read and search Slack freely. When given a message or thread link, read the whole thread, and the surrounding channel when context requires it.
- Never post, reply, react, edit, delete, forward, share, or change channel state without approval. Requests to draft, write, or improve a message mean draft only.
- Before a write, show the exact final content and action, then wait for an explicit instruction such as "send it". Approval is single-use and covers only that message.
- When unsure whether a write was authorized, do not perform it; ask.
