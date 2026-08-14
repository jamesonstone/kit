package templates

const awsAgentToolkitGuidanceRoute = `Before AWS-dependent work, load ` + "`docs/references/rules/aws-agent-toolkit-guidance.md`" + ` and use its current AWS skill, official documentation, AWS MCP Server or CLI fallback, identity, infrastructure-approval, and secret-safety routing; repo-local Kit gates remain authoritative.`

const memoryAWSContextHardGate = `## AWS Context Hard Gate

- ` + awsAgentToolkitGuidanceRoute + ` If ` + "`.kit.yaml`" + ` defines an enabled AWS context, run ` + "`kit aws verify`" + ` before the first AWS-dependent command and again immediately before AWS mutation
- Treat the verified account, ARN, and Region as authoritative; a profile name alone is not proof of identity
- Use the verified configured profile and Region explicitly for every AWS-dependent command where supported
- After verification, never use default, another discovered profile, or ambient credentials
- Stop on missing credentials, incomplete configuration, or identity mismatch

`
