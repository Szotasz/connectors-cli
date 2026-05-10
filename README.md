# connectors -- connectors.hu CLI

CLI for [connectors.hu](https://connectors.hu) -- Hungarian business API gateway (Billingo, NAV, MiniCRM).

## Install

```bash
curl -sL https://raw.githubusercontent.com/Szotasz/connectors-cli/main/install.sh | bash
```

Or build from source:

```bash
go install github.com/Szotasz/connectors-cli@latest
```

## Usage

```bash
export CONNECTORS_HU_TOKEN=cnk_your_api_key

connectors sync                                            # fetch tool manifest
connectors billingo list-documents --per_page 5            # list invoices
connectors nav query-taxpayer --taxNumber 12345678         # NAV taxpayer query
connectors billingo get-document --id 123 --select id,total  # field selection
```

## Claude Code integration

`connectors sync` auto-generates a Claude Code skill at `~/.claude/skills/connectors-hu/SKILL.md`, making all connector commands discoverable by Claude agents.

## Release

Tag a version to trigger binary builds:

```bash
git tag v0.2.0 && git push origin v0.2.0
```
