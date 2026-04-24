# midimap

A lightweight MIDI remapper scripted in Lua.

[日本語版 / Japanese README](./README-ja.md)

Route MIDI from one port to another through a Lua script that transforms notes, control changes, and other messages on the fly. Written in Go with the remapping logic fully delegated to Lua — edit your rules, restart, and you're done.

## Features

- Real-time MIDI remapping between any two MIDI ports
- Remapping rules written in Lua (notes, CCs, velocity scaling, key splits, etc.)
- Cross-platform: macOS (Apple Silicon / Intel), Windows (x64), Linux (x64 / arm64)
- Verbose mode with paired IN → OUT logging including raw hex bytes
- Lightweight single-binary deployment

## Quick Start

```bash
# 1. List available MIDI ports
midimap -l

# 2. Run the remapper
midimap -i 0 -o 1 -s luascripts/example.lua

# 3. With verbose logging
midimap -i 0 -o 1 -s luascripts/example.lua -v
```

## Installation

### macOS (Apple Silicon)

```bash
curl -LO https://github.com/mamemomonga/midimap/releases/latest/download/midimap-vX.Y.Z-darwin-arm64.tar.gz
tar xzf midimap-vX.Y.Z-darwin-arm64.tar.gz
cd midimap-vX.Y.Z-darwin-arm64
xattr -d com.apple.quarantine midimap
./midimap -l
```

### macOS (Intel)

```bash
curl -LO https://github.com/mamemomonga/midimap/releases/latest/download/midimap-vX.Y.Z-darwin-amd64.tar.gz
tar xzf midimap-vX.Y.Z-darwin-amd64.tar.gz
cd midimap-vX.Y.Z-darwin-amd64
xattr -d com.apple.quarantine midimap
./midimap -l
```

The `xattr` step removes the quarantine attribute that macOS attaches to unsigned downloaded binaries. Without it, Gatekeeper will block execution.

### Windows (x64)

PowerShell:

```powershell
Invoke-WebRequest -Uri https://github.com/mamemomonga/midimap/releases/latest/download/midimap-vX.Y.Z-windows-amd64.zip -OutFile midimap.zip
Expand-Archive midimap.zip -DestinationPath midimap
cd midimap\midimap-vX.Y.Z-windows-amd64
.\midimap.exe -l
```

On first run, Windows SmartScreen may warn about an unrecognized publisher. Click "More info" → "Run anyway".

### Linux (x64)

```bash
curl -LO https://github.com/mamemomonga/midimap/releases/latest/download/midimap-vX.Y.Z-linux-amd64.tar.gz
tar xzf midimap-vX.Y.Z-linux-amd64.tar.gz
cd midimap-vX.Y.Z-linux-amd64
./midimap -l
```

Requires ALSA runtime (installed by default on most desktop distributions). On minimal installs:

```bash
sudo apt-get install libasound2
```

### Linux (arm64)

```bash
curl -LO https://github.com/mamemomonga/midimap/releases/latest/download/midimap-vX.Y.Z-linux-arm64.tar.gz
tar xzf midimap-vX.Y.Z-linux-arm64.tar.gz
cd midimap-vX.Y.Z-linux-arm64
./midimap -l
```

Same ALSA requirement as x64.

## MIDI Port Setup

### macOS (IAC Driver)

macOS ships with a built-in virtual MIDI driver. Enable it once:

1. Open **Audio MIDI Setup** (`/System/Applications/Utilities/Audio MIDI Setup.app`)
2. Menu: **Window → Show MIDI Studio** (⌘2)
3. Double-click the **IAC Driver** icon
4. Check **"Device is online"**
5. Add two ports using the **+** button, e.g. `midimap-in` and `midimap-out`
6. Click **Apply**

Route your DAW or MIDI keyboard output to `midimap-in`, and listen for remapped MIDI on `midimap-out`.

### Windows (loopMIDI)

Windows has no built-in virtual MIDI driver. Install [loopMIDI](https://www.tobias-erichsen.de/software/loopmidi.html) (free):

1. Install and launch loopMIDI
2. Create two ports, e.g. `midimap-in` and `midimap-out`

### Linux (ALSA)

ALSA provides virtual MIDI ports through the `snd-virmidi` module:

```bash
sudo modprobe snd-virmidi
```

Or use JACK/Pipewire's built-in MIDI routing if already installed.

## Usage

```
midimap -l
midimap -i <in> -o <out> -s <script.lua> [-v]
```

### Options

| Flag | Description |
|------|-------------|
| `-i`, `-in <port>` | MIDI input port (name substring or number from `-l`) |
| `-o`, `-out <port>` | MIDI output port (name substring or number from `-l`) |
| `-l`, `-list` | List available MIDI ports and exit |
| `-s`, `-script <file>` | Lua remap script (required) |
| `-v`, `-verbose` | Print every MIDI event (IN and OUT) |
| `-V`, `-version` | Show version and exit |
| `-h`, `-help` | Show help |

### Examples

```bash
# List ports
midimap -l

# Connect by number
midimap -i 0 -o 1

# Connect by name (partial match)
midimap -i midimap-in -o midimap-out

# Verbose mode with a specific script
midimap -i 0 -o 1 -s myrules.lua -v
```

### Verbose Output Format

With `-v`, each MIDI event is shown as an IN → OUT pair with raw hex bytes:

```
NON C:00 N: 60 V:100(90 3C 64) -> NON C:00 N: 72 V:100(90 48 64)
CC  C:00 C:  1 V: 64(B0 01 40) -> CC  C:00 C: 11 V: 64(B0 0B 40)
```

Format:
- `NON` = Note On, `NOF` = Note Off, `CC` = Control Change
- `C:` = Channel, `N:` = Note, `V:` = Velocity/Value

## Writing Remap Rules

`midimap` calls Lua global functions for each MIDI event type. Define the ones you need; undefined events pass through unhandled (i.e. dropped).

### Minimal pass-through

```lua
function on_note_on(ch, note, vel)
    send_note_on(ch, note, vel)
end

function on_note_off(ch, note, vel)
    send_note_off(ch, note, vel)
end

function on_cc(ch, cc, val)
    send_cc(ch, cc, val)
end
```

### Callbacks you can define

| Function | Arguments |
|----------|-----------|
| `on_note_on(ch, note, vel)` | Channel, note number, velocity (0–127) |
| `on_note_off(ch, note, vel)` | Channel, note number, release velocity |
| `on_cc(ch, cc, val)` | Channel, CC number, value (0–127) |

### Send functions (callable from Lua)

| Function | Arguments |
|----------|-----------|
| `send_note_on(ch, note, vel)` | |
| `send_note_off(ch, note, vel)` | `vel` is accepted but ignored in output |
| `send_cc(ch, cc, val)` | |

### Channel numbering

MIDI channel values in Lua use **0–15** (not 1–16 as shown in DAWs).

| DAW display | Lua value |
|-------------|-----------|
| Ch 1        | 0         |
| Ch 2        | 1         |
| ...         | ...       |
| Ch 16       | 15        |

Note and CC numbers are 0–127 as usual.

### Example: transpose + CC remap

```lua
-- Transpose channel 0 up one octave
function on_note_on(ch, note, vel)
    if ch == 0 then
        send_note_on(ch, note + 12, vel)
    else
        send_note_on(ch, note, vel)
    end
end

function on_note_off(ch, note, vel)
    if ch == 0 then
        send_note_off(ch, note + 12, vel)
    else
        send_note_off(ch, note, vel)
    end
end

-- Remap mod wheel (CC1) to expression (CC11)
function on_cc(ch, cc, val)
    if cc == 1 then
        send_cc(ch, 11, val)
    else
        send_cc(ch, cc, val)
    end
end
```

See `luascripts/example.lua` for a working example.

## Building from Source

Requires Go 1.22+ and a C/C++ toolchain (cgo is used by the underlying RtMidi library).

```bash
git clone https://github.com/mamemomonga/midimap.git
cd midimap

# Build for your host
make build

# Run directly
make run IN=0 OUT=1

# With verbose
make dev IN=0 OUT=1

# All Make targets
make help
```

### Platform-specific build dependencies

- **macOS**: Xcode Command Line Tools (`xcode-select --install`)
- **Windows**: MSVC Build Tools or MinGW-w64
- **Linux**: `build-essential` and `libasound2-dev`
  ```bash
  sudo apt-get install build-essential libasound2-dev
  ```

### Cross-compilation

Cross-compiling with cgo is impractical across most platform boundaries. 

## Troubleshooting

**`midimap -l` shows no ports**
- macOS: enable IAC Driver in Audio MIDI Setup (see above)
- Windows: install loopMIDI
- Linux: `sudo modprobe snd-virmidi` or connect a physical MIDI device

**"permission denied" on first run (macOS / Linux)**
```bash
chmod +x midimap
```

**Gatekeeper blocks execution (macOS)**
```bash
xattr -d com.apple.quarantine midimap
```

**Note Off velocity shows 64 in verbose output**
This is expected. Many keyboards send Note Off as "Note On with velocity 0". The underlying MIDI library reports this as a Note Off with a placeholder velocity of 64. Compare the raw hex bytes to see the actual form on the wire (`90 ... 00` for running-status Note Off, `80 ... xx` for a true Note Off).

**Lua script errors print but don't crash**
Errors are logged to stderr; the remapper keeps running. Fix the script and restart.

## Dependencies

- [gomidi/midi](https://gitlab.com/gomidi/midi) — MIDI I/O (uses RtMidi under the hood)
- [yuin/gopher-lua](https://github.com/yuin/gopher-lua) — Lua 5.1 VM in pure Go

## License

MIT License. See [LICENSE](./LICENSE) for the full text.

## Contributing

Bug reports and pull requests welcome. For larger changes, please open an issue first to discuss what you'd like to change.

## Acknowledgments

Built on the excellent [RtMidi](https://www.music.mcgill.ca/~gary/rtmidi/) C++ library by Gary Scavone.

## Note

This was created using Claude Opus 4.7.