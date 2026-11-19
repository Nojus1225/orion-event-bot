# Orion Event Bot

Discord bot for Orion Hosting (orionhost.xyz) community events.

## Stack

- **Go** + [Disgo](https://github.com/disgoorg/disgo)
- **GORM** DB (SQLite)
- **go-i18n** locales

## Structure

- `main.go` — Entry point
- `bot/commands/` — Slash commands
- `bot/events/` — Discord listeners
- `bot/handlers/` — Command and event registration & syncing
- `bot/vars/` — Event configuration
- `db/` — GORM models
- `locales/` — Translation files

## Getting Started

1. Link the server with [Orion CLI](https://github.com/orionhosting/cli)
2. Build and run:

```bash
go build

./start.bash
```
