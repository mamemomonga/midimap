# midimap

Luaでリマッピングルールを記述する軽量MIDIリマッパー。

[English README](./README-en.md)

MIDIポート間で流れるノート、コントロールチェンジなどのメッセージを、Luaスクリプトでリアルタイムに変換します。本体はGoで書かれ、リマッピングのロジックはすべてLuaに委譲されています。ルールを編集して再起動するだけで反映されます。

## 特徴

- 任意の2つのMIDIポート間でリアルタイムにリマッピング
- リマッピングルールはLuaで記述(ノート、CC、ベロシティスケーリング、キースプリットなど)
- クロスプラットフォーム対応: macOS (Apple Silicon / Intel)、Windows (x64)、Linux (x64 / arm64)
- verboseモードでIN → OUTをペアで、RAWの16進バイトと共に表示
- シングルバイナリで配布

## クイックスタート

```bash
# 1. 利用可能なMIDIポート一覧を表示
midimap -l

# 2. リマッパーを起動
midimap -i 0 -o 1 -s luascripts/example.lua

# 3. verboseモードで起動
midimap -i 0 -o 1 -s luascripts/example.lua -v
```

## インストール

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

`xattr` コマンドは、macOSが未署名のダウンロードバイナリに付与する隔離属性を解除します。これを行わないとGatekeeperによって実行がブロックされます。

### Windows (x64)

PowerShell:

```powershell
Invoke-WebRequest -Uri https://github.com/mamemomonga/midimap/releases/latest/download/midimap-vX.Y.Z-windows-amd64.zip -OutFile midimap.zip
Expand-Archive midimap.zip -DestinationPath midimap
cd midimap\midimap-vX.Y.Z-windows-amd64
.\midimap.exe -l
```

初回起動時にWindows SmartScreenが「認識されていない発行元」と警告する場合があります。「詳細情報」→「実行」をクリックしてください。

### Linux (x64)

```bash
curl -LO https://github.com/mamemomonga/midimap/releases/latest/download/midimap-vX.Y.Z-linux-amd64.tar.gz
tar xzf midimap-vX.Y.Z-linux-amd64.tar.gz
cd midimap-vX.Y.Z-linux-amd64
./midimap -l
```

ALSAランタイムが必要です(多くのデスクトップディストリビューションにはデフォルトで入っています)。最小インストール環境では次のようにインストールしてください:

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

x64と同じくALSAが必要です。

## MIDIポートのセットアップ

### macOS (IAC Driver)

macOSには仮想MIDIドライバが標準搭載されています。最初に有効化するだけです:

1. **Audio MIDI設定** を開く (`/System/Applications/Utilities/Audio MIDI設定.app`)
2. メニュー: **ウィンドウ → MIDIスタジオを表示** (⌘2)
3. **IACドライバ** のアイコンをダブルクリック
4. **「装置はオンライン」** にチェック
5. **+** ボタンでポートを2つ追加(例: `midimap-in`、`midimap-out`)
6. **適用** をクリック

DAWやMIDIキーボードの出力を `midimap-in` に向け、リマップされたMIDIを `midimap-out` から受け取ります。

### Windows (loopMIDI)

Windowsには仮想MIDIドライバが標準搭載されていません。[loopMIDI](https://www.tobias-erichsen.de/software/loopmidi.html)(無料)をインストールしてください:

1. loopMIDIをインストールして起動
2. ポートを2つ作成(例: `midimap-in`、`midimap-out`)

### Linux (ALSA)

ALSAは `snd-virmidi` モジュールで仮想MIDIポートを提供します:

```bash
sudo modprobe snd-virmidi
```

JACKやPipewireを既に使っているなら、そちらの内蔵MIDIルーティングも利用できます。

## 使い方

```
midimap -l
midimap -i <in> -o <out> -s <script.lua> [-v]
```

### オプション

| フラグ | 説明 |
|--------|------|
| `-i`, `-in <ポート>` | MIDI入力ポート(名前の部分一致または `-l` の番号) |
| `-o`, `-out <ポート>` | MIDI出力ポート(名前の部分一致または `-l` の番号) |
| `-l`, `-list` | 利用可能なMIDIポートを一覧表示して終了 |
| `-s`, `-script <ファイル>` | Luaリマップスクリプト(必須) |
| `-v`, `-verbose` | すべてのMIDIイベントを表示 |
| `-V`, `-version` | バージョンを表示して終了 |
| `-h`, `-help` | ヘルプを表示 |

### 使用例

```bash
# ポート一覧
midimap -l

# 番号で指定
midimap -i 0 -o 1

# 名前で指定(部分一致)
midimap -i midimap-in -o midimap-out

# 指定スクリプトでverboseモード起動
midimap -i 0 -o 1 -s myrules.lua -v
```

### Verbose出力のフォーマット

`-v` を付けると、各MIDIイベントがIN → OUTのペアとしてRAWの16進バイトと共に表示されます:

```
NON C:00 N: 60 V:100(90 3C 64) -> NON C:00 N: 72 V:100(90 48 64)
CC  C:00 C:  1 V: 64(B0 01 40) -> CC  C:00 C: 11 V: 64(B0 0B 40)
```

フォーマット:
- `NON` = Note On、`NOF` = Note Off、`CC` = Control Change
- `C:` = チャンネル、`N:` = ノート、`V:` = ベロシティ/値

## リマップルールの書き方

`midimap` は各MIDIイベントタイプに対応するLuaグローバル関数を呼び出します。必要な関数だけを定義してください。未定義のイベントはそのまま落ちます(出力されません)。

### 最小限のパススルー

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

### 定義できるコールバック

| 関数 | 引数 |
|------|------|
| `on_note_on(ch, note, vel)` | チャンネル、ノート番号、ベロシティ (0〜127) |
| `on_note_off(ch, note, vel)` | チャンネル、ノート番号、リリースベロシティ |
| `on_cc(ch, cc, val)` | チャンネル、CC番号、値 (0〜127) |

### 送信関数(Luaから呼び出し可能)

| 関数 | 引数 |
|------|------|
| `send_note_on(ch, note, vel)` | |
| `send_note_off(ch, note, vel)` | `vel` は受け取るが出力では無視される |
| `send_cc(ch, cc, val)` | |

### チャンネル番号について

LuaスクリプトでのMIDIチャンネル値は **0〜15** を使います(DAW画面上の表記 1〜16 ではありません)。

| DAW上の表記 | Luaでの値 |
|-------------|-----------|
| Ch 1        | 0         |
| Ch 2        | 1         |
| ...         | ...       |
| Ch 16       | 15        |

ノート番号・CC番号は従来通り 0〜127 です。

### 例: 移調 + CCリマップ

```lua
-- チャンネル0を1オクターブ上げる
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

-- モジュレーションホイール(CC1)をエクスプレッション(CC11)にリマップ
function on_cc(ch, cc, val)
    if cc == 1 then
        send_cc(ch, 11, val)
    else
        send_cc(ch, cc, val)
    end
end
```

動作サンプルは `luascripts/example.lua` を参照してください。

## ソースからビルド

Go 1.22以上と、C/C++ツールチェーンが必要です(内部のRtMidiライブラリがcgoを使用するため)。

```bash
git clone https://github.com/mamemomonga/midimap.git
cd midimap

# ホスト向けにビルド
make build

# そのまま実行
make run IN=0 OUT=1

# verboseで実行
make dev IN=0 OUT=1

# Makeターゲット一覧
make help
```

### プラットフォーム別のビルド依存

- **macOS**: Xcodeコマンドラインツール (`xcode-select --install`)
- **Windows**: MSVC Build ToolsまたはMinGW-w64
- **Linux**: `build-essential` と `libasound2-dev`
  ```bash
  sudo apt-get install build-essential libasound2-dev
  ```

### クロスコンパイル

cgoを使うプロジェクトのクロスコンパイルはOSを跨ぐと現実的ではありません。

## トラブルシューティング

### `midimap -l` でポートが表示されない
- macOS: Audio MIDI設定でIACドライバを有効化(上記参照)
- Windows: loopMIDIをインストール
- Linux: `sudo modprobe snd-virmidi` または物理MIDIデバイスを接続

### 初回起動時に「permission denied」(macOS / Linux)
```bash
chmod +x midimap
```

### Gatekeeperで実行がブロックされる(macOS)
```bash
xattr -d com.apple.quarantine midimap
```

### Verbose出力でNote Offのベロシティが64と表示される
これは仕様どおりの挙動です。多くのキーボードはNote Offを「ベロシティ0のNote On」として送信します。内部で使っているMIDIライブラリはこれをNote Offとして報告しますが、リリースベロシティが存在しないためダミー値として64が入ります。実際のバイト列を確認するにはRAWの16進表示を見てください(`90 ... 00` ならrunning-status形式のNote Off、`80 ... xx` なら本物のNote Off)。

### Luaスクリプトのエラーが表示されるがクラッシュしない
エラーはstderrに出力され、リマッパーは動作を継続します。スクリプトを修正して再起動してください。

## ドキュメントの翻訳方針

- 日本語がプライマリ、英語は自動生成
  - `CLAUDE.md` ⇒ `CLAUDE-en.md`(Claude Code は参照しない)
  - `README.md` ⇒ `README-en.md`
- 翻訳ルール:
  - 技術用語 (cgo, goroutine, MIDI, CC 等) は英語のまま
  - コードブロック・ファイルパス・コマンド例は改変しない
  - Markdown構造(見出し、リスト、表)は完全維持
  - 自然な技術英語で、冗長にしない
- 英語版を手で編集しないこと(次回翻訳で上書きされる)
- 翻訳の実行: `make translate` または セッションで `/translate-docs`

## 依存ライブラリ

- [gomidi/midi](https://gitlab.com/gomidi/midi) — MIDI I/O(内部でRtMidiを使用)
- [yuin/gopher-lua](https://github.com/yuin/gopher-lua) — 純GoのLua 5.1 VM

## ライセンス

MITライセンス。詳細は [LICENSE](./LICENSE) を参照してください。

## コントリビューション

バグ報告やPull Requestを歓迎します。大きな変更を行う場合は、まずIssueで相談してください。

## 謝辞

Gary Scavone氏による優れたC++ライブラリ [RtMidi](https://www.music.mcgill.ca/~gary/rtmidi/) の上に構築されています。

## 備考

制作には Claude Opus 4.7を使用しています。