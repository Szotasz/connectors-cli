# conn -- connectors.hu CLI

CLI for [connectors.hu](https://connectors.hu) -- Hungarian business API gateway (Billingo, NAV, MiniCRM).

## Install

```bash
curl -sL https://raw.githubusercontent.com/Szotasz/conn-cli/main/install.sh | bash
```

Or build from source:

```bash
go install github.com/Szotasz/conn-cli@latest
```

## Usage

```bash
export CONN_HU_TOKEN=cnk_your_api_key

conn sync                                        # fetch tool manifest
conn billingo list-documents --per_page 5         # list invoices
conn nav query-taxpayer --taxNumber 12345678      # NAV taxpayer query
conn billingo get-document --id 123 --select id,total  # field selection
```

## Claude Code integration

`conn sync` auto-generates a Claude Code skill at `~/.claude/skills/conn-hu/SKILL.md`, making all connector commands discoverable by Claude agents.

## Release

Tag a version to trigger binary builds:

```bash
git tag v0.1.0 && git push origin v0.1.0
```
