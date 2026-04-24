# midimap - Claude Code 向けプロジェクト文脈

## 概要

Luaスクリプトでリマッピングルールを記述する MIDIリマッパー。
本体は Go、リマップロジックは Lua。MIDI I/O は gomidi/midi/v2
(RtMidi ベース、cgo 必須)。

## 構成

- `main.go` - CLI エントリ、フラグ解析、シグナルハンドリング
- `bridge.go` - Lua ⇔ Go ブリッジ: dispatch、callLua、registerAPI、formatMsg
- `luascripts/` - ユーザー向け Lua リマップスクリプト(自動読み込みなし、-s は必須)
  - `example.lua` - 最小サンプル(移調・ベロシティ半減・CC付け替え)
  - `thru.lua` - 入力をそのまま出力に流すパススルー
  - `L6max_Mono.lua` - ZOOM LiveTrak L6max向け、チャンネルコントロール(8ch独立)
  - `L6max_Stereo.lua` - 同上のch1,2 および ch3,4のステレオリンク版

## 重要な設計判断

- MIDIチャンネル値は 0〜15(DAW表示の1〜16ではない)。
  例: `send_cc(0, ...)` は MIDI Ch 1 を指す。
- `-s` は必須。デフォルトスクリプトなし。暗黙のパススルーより
  「指定漏れはエラー」を優先する。
- `send_note_off(ch, note, vel)` の vel は受け取るが破棄される。
  これは gomidi の非対称な API に合わせた仕様。対称化しないこと。
- NoteOn vel=0 は gomidi が NoteOff として解釈し、verbose ログに
  placeholder velocity 64 として表示される。これは README に
  明記済みの仕様。dispatch ロジックを変更して隠蔽しないこと。
- 未処理のMIDIメッセージ種別(SysEx、Pitch Bend、Aftertouch)は
  現状捨てている。拡張する場合は bridge.go の `dispatch()` に追加する。
- gopher-lua はシングルスレッド。MIDIリスナーは sync.Mutex で
  直列化している。Lua 呼び出しに goroutine による並列化を入れないこと。

## Verboseログのフォーマット

IN と OUT イベントを「ラウンド」単位でペアリング(1つのIN、0個以上のOUT)。
フォーマット: `NON C:cc N:nn V:vv(HH HH HH) -> NON C:cc N:nn V:vv(HH HH HH)`
- 数値はスペース埋め (`%3d`、`%03d` ではない) で可読性優先
- チャンネルもスペース埋め (`%2d`)、桁を揃えて可読性優先
- RAWバイトは大文字16進、スペース区切り
詳細は bridge.go の `formatMsg` と `eventLogger` を参照。

## ビルドとリリース

- ローカル: `make build` (cgo ツールチェーン必須)
- クロスプラットフォームリリースは GitHub Actions: `v*` タグで
  .github/workflows/release.yml が起動
- macos-14 (Apple Silicon) で `-target x86_64-apple-macos11` を
  使って darwin-amd64 をクロスビルド。macos-13 はキュー時間と
  廃止予定のため使っていない。
- Linux arm64 は ubuntu-22.04-arm(パブリックリポジトリでは無料)

## 依存ライブラリ

- gitlab.com/gomidi/midi/v2 (+ rtmididrv) - MIDI I/O、cgo必須
  - NoteOn: (ch, key, vel) - 3引数
  - NoteOff: (ch, key) - 2引数(velocity なし)
  - GetNoteOn / GetNoteOff / GetControlChange - すべて3ポインタ引数
- github.com/yuin/gopher-lua - Lua 5.1 VM、純Go、cgo 不要

## コーディング規約

- Go のエラーは `error` として伝播、プロセス終了は main.go の log.Fatal のみ
- Lua エラーはログに出すがプロセスは落とさない (Protect: true)
- cgo を必要とする依存を新たに追加しないこと(どうしても必要な場合を除く)
- README.md (英語) と README-ja.md (日本語) は常に同期させる

## やってはいけないこと

- `-s` をオプショナルにしてデフォルト値を持たせる
- Note Off velocity=64 の表示を「正規化」する
- require() のパス解決を再導入する(検討の上、除外済み)
- Lua コールバックに並列実行を持ち込む
