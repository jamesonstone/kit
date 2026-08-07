package templates

func codexBrowserPolicy(title string) string {
	if title != "AGENTS" {
		return ""
	}

	return `## Browser policy

- For interactive browser work, use Codex's built-in browser through ` + "`@Browser`" + `.
- Do not use ` + "`@Chrome`" + `, control my active Chrome profile, or launch external
  Chrome or Chromium through Playwright, Selenium, Cypress, or browser MCP tools
  unless I explicitly request it.
- If ` + "`@Browser`" + ` is unavailable, report the limitation instead of silently
  falling back.
- When I explicitly authorize an external browser, terminate and verify all
  task-owned browser and automation processes before finishing.

`
}
