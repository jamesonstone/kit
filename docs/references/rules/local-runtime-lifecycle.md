---
kind: ruleset
slug: local-runtime-lifecycle
description: Keeps one local runtime per project, reuses or stops it before any start, restarts only what changed, and keeps agents from leaving terminal windows behind.
status: active
registry_scope: downstream
applies_to:
  - local-environment
  - rungrid
  - terminal
  - dev-server
  - validation
  - cleanup
read_policy_default: conditional
---

# Ruleset: local-runtime-lifecycle

## Purpose

- Keep one local runtime per project on the machine, so repeated starts cannot exhaust the CPU, the GPU, memory, or ports.
- Keep routine agent restarts from leaving terminal windows, tabs, and processes behind for the human to find.

## Applies When

- Starting, restarting, or stopping a local environment: a workspace manager such as rungrid, a Compose stack, a dev server, a watcher, or a worker.
- Opening a terminal window or tab for a local environment, or running a validation step or post-merge check that needs one.

## Rules

### Inspect Before Starting

- Before any start, find what already runs for this project: the manager's status (`rungrid status`), the project's listening ports, and its processes. Do this every time, including inside a loop.
- Run one runtime per project. If it is healthy, reuse it and start nothing. Do not start a second copy for a worktree; use the manager's override instead (`rungrid up --override` or `rungrid <service> <worktree>`).
- If an old or broken instance of this project blocks the work, stop it through its manager, confirm that it stopped, and then start one replacement. Never start beside it.
- If another session or the human is using the instance, do not stop it. Reuse it, or report the conflict.

### Restart Only What Changed

- Restart the affected service (`rungrid stop <service>`, then `rungrid start <service>`). Use a full stop and start only when the change needs it, and report why.
- Do not restart to check health; read the health endpoints. A merge that changes no running code needs no restart.

### No Terminal Windows From Agents

- Start without opening terminal windows or tabs: `rungrid up --no-open`, or the manager's equivalent. Use a window-opening start, `rungrid open`, `rungrid resume`, or `rungrid attach` only when the human asks, and then only once.
- A plain `rungrid up` opens a new window with every tab each time, even when it reuses a running runtime, and `rungrid down` closes none of them. Never run it in a loop, in a retry, or as a post-merge step.
- Read service output with the manager's log command (`rungrid logs <service>`) or redirected log files, not with an attached terminal.

### Verify Teardown

- After a stop, confirm that the manager reports it stopped, the ports are free, and its processes exited. Report anything left running.
- When the work ends, stop the processes, dev servers, and watchers that it started, under `work-cleanup`. Leave a runtime that was already running as you found it.
- Stop processes only through their manager or by exact process ID. Never kill by a broad name pattern, and never close a terminal window or tab that you did not open.
- If your work opened windows that you cannot close, tell the human how many, and stop opening more.
