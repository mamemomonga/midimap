package main

import (
	"fmt"
	"strings"

	lua "github.com/yuin/gopher-lua"
	"gitlab.com/gomidi/midi/v2"
)

// verbose 用のイベントロガー。IN と OUT をラウンドでペアにする
type eventLogger struct {
	enabled bool
	in      string   // 現在ラウンドの IN 文字列(空なら IN なし)
	outs    []string // 現在ラウンドの OUT 文字列
	inRound bool     // ラウンド内かどうか
}

func (e *eventLogger) beginRound(in midi.Message) {
	if !e.enabled {
		return
	}
	e.in = formatMsg(in)
	e.outs = e.outs[:0]
	e.inRound = true
}

func (e *eventLogger) recordOut(out midi.Message) {
	if !e.enabled {
		return
	}
	if e.inRound {
		e.outs = append(e.outs, formatMsg(out))
	} else {
		// ラウンド外(IN なしの OUT): その場で出力
		fmt.Printf("%s -> %s\n", strings.Repeat(" ", msgFieldWidth()), formatMsg(out))
	}
}

func (e *eventLogger) endRound() {
	if !e.enabled || !e.inRound {
		return
	}
	switch len(e.outs) {
	case 0:
		// 入力のみ(フィルタされた)
		fmt.Printf("%s -> \n", e.in)
	case 1:
		fmt.Printf("%s -> %s\n", e.in, e.outs[0])
	default:
		// 1対多: 最初の行に IN + 最初の OUT、以降はインデントして OUT のみ
		fmt.Printf("%s -> %s\n", e.in, e.outs[0])
		pad := strings.Repeat(" ", len(e.in))
		for _, o := range e.outs[1:] {
			fmt.Printf("%s -> %s\n", pad, o)
		}
	}
	e.in = ""
	e.outs = e.outs[:0]
	e.inRound = false
}

// IN なし OUT の左側パディング幅(おおよその整列用)
func msgFieldWidth() int {
	// "NON C:00 N:060 V:100(90 3C 64)" の長さ
	return 30
}

// メッセージを1つの文字列に整形
func formatMsg(msg midi.Message) string {
	var ch, k, v uint8
	switch {
	case msg.GetNoteOn(&ch, &k, &v):
		return fmt.Sprintf("NON C:%2d N:%3d V:%3d(%s)", ch, k, v, rawHex(msg))
	case msg.GetNoteOff(&ch, &k, &v):
		return fmt.Sprintf("NOF C:%2d N:%3d V:%3d(%s)", ch, k, v, rawHex(msg))
	case msg.GetControlChange(&ch, &k, &v):
		return fmt.Sprintf("CC  C:%2d C:%3d V:%3d(%s)", ch, k, v, rawHex(msg))
	default:
		return fmt.Sprintf("%s(%s)", msg.String(), rawHex(msg))
	}
}

// 生バイトを "90 3C 64" 形式に (大文字2桁16進)
func rawHex(msg midi.Message) string {
	b := msg.Bytes()
	parts := make([]string, len(b))
	for i, x := range b {
		parts[i] = fmt.Sprintf("%02X", x)
	}
	return strings.Join(parts, " ")
}

// Goが受信したMIDIメッセージを対応するLua関数へディスパッチ
func dispatch(L *lua.LState, msg midi.Message) {
	var ch, k, v uint8
	switch {
	case msg.GetNoteOn(&ch, &k, &v):
		callLua(L, "on_note_on", float64(ch), float64(k), float64(v))
	case msg.GetNoteOff(&ch, &k, &v):
		callLua(L, "on_note_off", float64(ch), float64(k), float64(v))
	case msg.GetControlChange(&ch, &k, &v):
		callLua(L, "on_cc", float64(ch), float64(k), float64(v))
	}
}

func callLua(L *lua.LState, fn string, args ...float64) {
	f := L.GetGlobal(fn)
	if f.Type() == lua.LTNil {
		return
	}
	largs := make([]lua.LValue, len(args))
	for i, a := range args {
		largs[i] = lua.LNumber(a)
	}
	if err := L.CallByParam(lua.P{Fn: f, NRet: 0, Protect: true}, largs...); err != nil {
		fmt.Println("lua error:", err.Error())
	}
}

// Luaから呼び出せるMIDI送信API
func registerAPI(L *lua.LState, send func(midi.Message) error) {
	send3 := func(build func(ch, k, v uint8) midi.Message) lua.LGFunction {
		return func(L *lua.LState) int {
			ch := uint8(L.CheckInt(1))
			k := uint8(L.CheckInt(2))
			v := uint8(L.CheckInt(3))
			if err := send(build(ch, k, v)); err != nil {
				L.RaiseError("send failed: %v", err)
			}
			return 0
		}
	}
	send2 := func(build func(ch, k uint8) midi.Message) lua.LGFunction {
		return func(L *lua.LState) int {
			ch := uint8(L.CheckInt(1))
			k := uint8(L.CheckInt(2))
			if err := send(build(ch, k)); err != nil {
				L.RaiseError("send failed: %v", err)
			}
			return 0
		}
	}

	L.SetGlobal("send_note_on", L.NewFunction(send3(midi.NoteOn)))
	L.SetGlobal("send_note_off", L.NewFunction(send2(midi.NoteOff)))
	L.SetGlobal("send_cc", L.NewFunction(send3(midi.ControlChange)))
}
