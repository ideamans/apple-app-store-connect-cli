# Publishing the apple-app-store-connect-cli plugin

## Before every release

1. `go generate ./...` — regenerate `internal/llmdocs/90-commands.md`, commit
   any diff.
2. `go test ./...` — includes `TestPluginSkills`, which enforces that
   `plugin.json.version` equals `PluginVersion` in `cmd/root.go` and that the
   SKILL.md frontmatter stays within the Agent Skills standard.
3. `claude plugin validate plugins/apple-app-store-connect-cli`.
4. Bump `PluginVersion` and `plugin.json.version` together, in the same commit
   as the release tag. The release workflow refuses a mismatched tag.

## Registering in the marketplace (first release only)

Add to `.claude-plugin/marketplace.json` in `ideamans/claude-public-plugins`:

```json
{
  "name": "apple-app-store-connect-cli",
  "source": {
    "source": "git-subdir",
    "url": "https://github.com/ideamans/apple-app-store-connect-cli.git",
    "path": "plugins/apple-app-store-connect-cli"
  }
}
```

## Verifying the published result

```
/plugin marketplace add ideamans/claude-public-plugins
/plugin install apple-app-store-connect-cli@ideamans-plugins
/asc-usage
```

Other hosts install the same files directly:

```bash
gh skill install ideamans/apple-app-store-connect-cli/plugins/apple-app-store-connect-cli/skills/asc-usage --agent copilot
```
