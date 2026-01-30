# 🛎️ Faustus

> A beautiful TUI for managing Claude Code sessions

[![asciicast](./assets/demo.gif)](https://asciinema.org/a/FApqeZY0R9kWzR07)

## Features

- **Browse Sessions**: View all your Claude Code conversation sessions
- **Filter**: Filter session list by summary, prompt, project name
- **Deep Search**: Search through all session content (messages, code, etc.)
- **Preview Pane**: View conversation content with search highlighting
- **Delete**: Move sessions to bin (recoverable)
- **Restore**: Recover sessions from bin
- **Rename**: Update session summaries
- **Bin Management**: Empty bin to permanently delete sessions

## Installation

```bash
go install github.com/Fuwn/faustus@latest
```

Or build from source:

```bash
git clone https://github.com/Fuwn/faustus.git
cd faustus
task build
task install
```

## Usage

```bash
faustus
```

## Keybindings

Vim-style navigation:

| Key | Action |
|-----|--------|
| `j/k` | Navigate down/up (or scroll preview when focused) |
| `h/l` | Switch tabs (Sessions ↔ Bin) |
| `gg/G` | Jump to top/bottom |
| `C-u/C-d` | Half page up/down |
| `/` | Filter list (or search in preview when focused) |
| `s` | Deep search across all session content |
| `n/N` | Next/previous search match |
| `p` | Toggle preview pane |
| `tab` | Switch focus between list and preview |
| `d` | Delete (move to bin) |
| `u` | Restore from bin |
| `c` | Change name (rename) |
| `D` | Clear bin |
| `?` | Toggle help |
| `q` | Quit |

## Search

- **Filter (`/`)**: Filters the session list by summary, first prompt, and project name. When the preview is focused, searches within the current preview.
- **Deep Search (`s`)**: Searches through all message content across all sessions. Results show context around matches. Use `n/N` to navigate between matches.

## Data Location

Sessions are stored in `~/.claude/projects/`. Binned sessions are moved to `~/.claude/faustus-trash/`.

## Licence

Licensed under either of [Apache License, Version 2.0](LICENSE-APACHE) or
[MIT license](LICENSE-MIT) at your option.

Unless you explicitly state otherwise, any contribution intentionally submitted
for inclusion in this crate by you, as defined in the Apache-2.0 license, shall
be dual licensed as above, without any additional terms or conditions.
