package main

import (
	"gitlab.com/gomidi/midi/v2"
	lua "github.com/yuin/gopher-lua"
)

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
		return // そのイベント用のハンドラが未定義ならスルー
	}
	largs := make([]lua.LValue, len(args))
	for i, a := range args {
		largs[i] = lua.LNumber(a)
	}
	if err := L.CallByParam(lua.P{Fn: f, NRet: 0, Protect: true}, largs...); err != nil {
		// エラーでプロセスを落とさず、ログだけ出す
		println("lua error:", err.Error())
	}
}

// Luaから呼び出せるMIDI送信API
func registerAPI(L *lua.LState, send func(midi.Message) error) {
	// 3引数 (ch, key, val) 用 — NoteOn, ControlChange
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

	// 2引数 (ch, key) 用 — NoteOff
	send2 := func(build func(ch, k uint8) midi.Message) lua.LGFunction {
		return func(L *lua.LState) int {
			ch := uint8(L.CheckInt(1))
			k := uint8(L.CheckInt(2))
			// Lua側が3引数で呼んでも第3引数は無視
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