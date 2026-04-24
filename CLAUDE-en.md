# midimap - Project Context for Claude Code

## Overview

A MIDI remapper whose mapping rules are written in Lua scripts.
The core is Go; the remap logic is Lua. MIDI I/O uses gomidi/midi/v2
(RtMidi-based, cgo required).

## Structure

- `main.go` - CLI entry, flag parsing, signal handling
- `bridge.go` - Lua ⇔ Go bridge: dispatch, callLua, registerAPI, formatMsg
- `luascripts/` - user-facing Lua remap scripts (no auto-loading; -s is required)
  - `example.lua` - minimal sample (transpose, halved velocity, CC remap)
  - `thru.lua` - pure pass-through from input to output
  - `L6max_Mono.lua` - channel control for ZOOM LiveTrak L6max (8 independent channels)
  - `L6max_Stereo.lua` - same as above with ch1,2 and ch3,4 stereo-linked

## Key Design Decisions

- MIDI channel values are 0–15 (not the 1–16 shown in DAWs).
  e.g. `send_cc(0, ...)` refers to MIDI Ch 1.
- `-s` is required. No default script. We prefer "missing spec is an error"
  over implicit pass-through.
- The `vel` argument of `send_note_off(ch, note, vel)` is accepted but
  discarded. This mirrors gomidi's asymmetric API. Do not symmetrize it.
- NoteOn vel=0 is interpreted as NoteOff by gomidi and shown in verbose
  logs with a placeholder velocity of 64. This is documented behavior in
  README. Do not change dispatch logic to hide it.
- Unhandled MIDI message types (SysEx, Pitch Bend, Aftertouch) are
  currently dropped. To extend, add cases to `dispatch()` in bridge.go.
- gopher-lua is single-threaded. The MIDI listener serializes callbacks
  via sync.Mutex. Do not introduce goroutine parallelism into Lua calls.

## Verbose Log Format

IN and OUT events are paired per "round" (one IN, zero or more OUTs).
Format: `NON C:cc N:nn V:vv(HH HH HH) -> NON C:cc N:nn V:vv(HH HH HH)`
- Numbers use space-padding (`%3d`, not `%03d`) for readability
- Channel is also space-padded (`%2d`) to align columns for readability
- Raw bytes are uppercase hex, space-separated
See `formatMsg` and `eventLogger` in bridge.go for details.

## Build and Release

- Local: `make build` (cgo toolchain required)
- Cross-platform releases run on GitHub Actions: a `v*` tag triggers
  .github/workflows/release.yml
- darwin-amd64 is cross-built on macos-14 (Apple Silicon) using
  `-target x86_64-apple-macos11`. macos-13 is avoided due to queue
  times and upcoming deprecation.
- Linux arm64 uses ubuntu-22.04-arm (free for public repos)

## Dependencies

- gitlab.com/gomidi/midi/v2 (+ rtmididrv) - MIDI I/O, cgo required
  - NoteOn: (ch, key, vel) - 3 args
  - NoteOff: (ch, key) - 2 args (no velocity)
  - GetNoteOn / GetNoteOff / GetControlChange - all take 3 pointer args
- github.com/yuin/gopher-lua - Lua 5.1 VM, pure Go, no cgo

## Coding Conventions

- Go errors propagate as `error`; process exit only via log.Fatal in main.go
- Lua errors are logged but do not crash the process (Protect: true)
- Do not add new cgo-dependent dependencies (unless strictly necessary)
- README.md (Japanese) and README-en.md (English) must always stay in sync

## Documentation Translation Policy

- Japanese is the primary source; English is auto-generated
  - `CLAUDE.md` ⇒ `CLAUDE-en.md` (Claude Code does not read this)
  - `README.md` ⇒ `README-en.md`
- Translation rules:
  - Keep technical terms (cgo, goroutine, MIDI, CC, etc.) in English
  - Do not alter code blocks, file paths, or command examples
  - Preserve Markdown structure (headings, lists, tables) exactly
  - Use natural technical English; avoid verbosity
- Do not hand-edit the English version (it will be overwritten by the next translation)
- To run the translation: `make translate` or `/translate-docs` in a session

## Do Not

- Make `-s` optional with a default value
- "Normalize" the Note Off velocity=64 display
- Reintroduce require() path resolution (considered and rejected)
- Introduce parallel execution in Lua callbacks
