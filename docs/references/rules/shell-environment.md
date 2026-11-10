---
kind: ruleset
slug: shell-environment
description: Resolves project and user shell initialization without sharing personal configuration or bypassing execution policy.
status: active
registry_scope: downstream
applies_to:
  - shell
  - environment
  - toolchain
  - validation
read_policy_default: conditional
---

# Ruleset: shell-environment

## Purpose

Run shell-dependent commands in the intended project environment for the current user and executor. Keep personal configuration out of shared instructions.

## Applies When

- Before a command depends on shell startup files, PATH, aliases, functions, a version manager, or project activation.
- When a command differs from the user's terminal, after an executor or host update, or when a required tool is missing.

## Rules

### Resolve the environment

- Read `.kit.yaml` `shell.command` first. Omitted, null, or empty configuration means the current user's initialized interactive-login shell environment. Resolve it on the current executor from the user's approved host configuration; do not freeze one user's shell into shared config. Explicit script, container, or CI contracts still take precedence for those commands.
- A nonempty `shell.command` is an exact argument-vector prefix, for example `[zsh, -lic]`. It contains the shell executable and options; append the task command as one argument through the host's supported execution tool. Use all configured options as written. Do not evaluate the prefix as shell code or expand `$SHELL`, `~`, or other substitutions in it. A missing executable or unsupported host capability is a blocker, not a reason to ignore the override.
- Validate configuration with `kit config check`. Kit reads, validates, and preserves this preference; it does not launch a shell, alter its own process environment, or force an execution host to honor it. Review an unfamiliar prefix under normal execution permissions before use; configuration cannot grant permission to run arbitrary startup code.

- Read project instructions and the Environment Preflights section of `docs/references/testing.md`. Follow its link to existing environment documentation when present. Inspect the relevant toolchain manifest, script interpreter, task runner, and activation procedure before choosing an invocation.
- Keep that project-owned reference current with the working directory, supported hosts/shells or container, toolchain versions, approved activation procedure, required variable names without values, and harmless verification commands. Record local/CI differences. Reuse existing documentation instead of creating competing instructions.
- Use each team member's own approved host configuration. Keep personal shell paths, startup-file locations, aliases, functions, account selection, and machine preferences in that user's global agent instructions or untracked local configuration. Never commit another user's home path, startup contents, credentials, or a captured environment.
- `$SHELL` is a preference hint, not proof of the running interpreter or startup mode. Verify the current executor and available shell. Do not assume a desktop, container, remote host, and CI job share a home directory or shell.

### Initialize what the command requires

- Prefer the project's explicit execution contract: its script shebang, task runner, container, or noninteractive CI setup. Preserve each script's declared interpreter. Use a user shell as a parent wrapper only when the command needs that user's initialized environment.
- When local commands require interactive login configuration, request both modes through the host tool or one invocation of the verified shell. A `login=true` setting alone does not prove interactive startup. For zsh, `zsh -lic '<command>'` requests both modes. For Bash, use `bash -lic '<command>'` only when the user's startup convention supports it: Bash login startup does not automatically read `.bashrc`; an existing profile may source it. For fish, PowerShell, or other shells, use their documented startup mechanism rather than copying POSIX flags.
- Zsh normally reads `.zshenv`, `.zprofile` for login shells, `.zshrc` for interactive shells, and `.zlogin` for login shells, with corresponding system files and `ZDOTDIR` behavior. Use the live files through normal startup. Do not source them twice, freeze PATH values, or copy startup contents or credentials into instructions.
- Set the intended project/worktree working directory explicitly. Apply project activation through its approved workflow. An interactive non-terminal invocation may not run prompt hooks, including automatic environment activation. Do not silently approve `.envrc`, install a manager, or edit startup files to make a command work.
- Invoke shell functions inside the initialized shell that defines them. Functions and aliases are not portable environment variables; child processes inherit only exported variables, not every shell variable or function. Never extract, print, or copy definitions or credential values to another process to simulate initialization.
- Avoid nested wrappers when the current invocation already has the required modes and environment. Keep hermetic commands and CI noninteractive when personal interactive configuration is unnecessary. Pass arguments safely; never interpolate untrusted text into shell code.

### Verify without exposing secrets

- Use harmless probes in the same invocation as the affected command when initialization matters: shell mode, `command -v <tool>`, and tool version. A zsh mode probe is `[[ -o login && -o interactive ]]`; use shell-native equivalents elsewhere. Inspect required variable presence, never its value.
- Never dump `env`, `printenv`, shell state, startup-file contents, or credentials. Inspect only facts required for the task.
- Reverify after a host update, executor change, or reported mismatch. Distinguish observed behavior from assumptions; instructions do not guarantee that every later session obeys them.

### Keep permission failures separate

- Shell initialization cannot grant network, sandbox, filesystem, authentication, or service permissions. Missing executables, wrong versions, missing activation, authentication failures, and policy denials need different remedies.
- After an explicit policy denial, stop the denied action. Do not change proxies, network/security settings, credentials, endpoints, executors, or tools to bypass it. Report the blocked action and boundary, then seek the required authorization or supported setup. Continue independent allowed work.
- If required startup modes or activation are unavailable, report the missing capability and checks that did not run. A shell name or unrelated successful command does not establish environment readiness.

## Verification

- Commands use the project-owned environment contract and the current user's approved host configuration.
- Safe probes establish required shell modes, executable resolution, and versions without secret output.
- Scripts and CI retain their declared execution environment; permission failures remain explicit.
- Init and reconcile install the rule and routing while preserving project customizations. These checks verify shipped guidance, not compliance by all agent hosts.
