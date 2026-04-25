package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	lua "github.com/yuin/gopher-lua"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

var version = "dev"

const usage = `midimap - MIDI remapper scripted in Lua

USAGE:
  midimap -l
  midimap -i <in> -o <out> -s <script.lua> [-v]

OPTIONS:
  -i, -in <port>       MIDI input port (name substring or number from -l)
  -o, -out <port>      MIDI output port (name substring or number from -l)
  -s, -script <file>   Lua remap script (required)
  -l, -list            List available MIDI ports and exit
  -v, -verbose         Print every MIDI event (in and out)
  -V, -version         Show version and exit
  -h, -help            Show this help

EXAMPLES:
  midimap -l
  midimap -i 0 -o 1 -s luascripts/example.lua
  midimap -i midimap-in -o midimap-out -s luascripts/mixer.lua -v
`

// 片方を長短どちらで書いても同じ変数を共有するフラグ
type stringFlag struct {
	short, long string
	value       string
}

func (s *stringFlag) register(fs *flag.FlagSet, defaultVal, desc string) {
	s.value = defaultVal
	fs.StringVar(&s.value, s.short, defaultVal, desc)
	fs.StringVar(&s.value, s.long, defaultVal, desc)
}

type boolFlag struct {
	short, long string
	value       bool
}

func (b *boolFlag) register(fs *flag.FlagSet, desc string) {
	fs.BoolVar(&b.value, b.short, false, desc)
	fs.BoolVar(&b.value, b.long, false, desc)
}

func main() {
	fs := flag.NewFlagSet("midimap", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }

	inFlag := stringFlag{short: "i", long: "in"}
	outFlag := stringFlag{short: "o", long: "out"}
	scriptFlag := stringFlag{short: "s", long: "script"}
	listFlag := boolFlag{short: "l", long: "list"}
	verboseFlag := boolFlag{short: "v", long: "verbose"}
	versionFlag := boolFlag{short: "V", long: "version"}

	inFlag.register(fs, "", "MIDI input port (name or number)")
	outFlag.register(fs, "", "MIDI output port (name or number)")
	scriptFlag.register(fs, "", "Lua remap script (required)")
	listFlag.register(fs, "list available MIDI ports")
	verboseFlag.register(fs, "print every MIDI event")
	versionFlag.register(fs, "show version and exit")

	// 引数なしならヘルプを出して終了
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(0)
	}

	if err := fs.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		os.Exit(2)
	}

	if versionFlag.value {
		fmt.Printf("midimap %s\n", version)
		return
	}

	if listFlag.value {
		listPorts()
		return
	}

	if inFlag.value == "" || outFlag.value == "" || scriptFlag.value == "" {
		fmt.Fprintln(os.Stderr, "error: -in, -out, and -script are required")
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	if err := run(inFlag.value, outFlag.value, scriptFlag.value, verboseFlag.value); err != nil {
		log.Fatal(err)
	}
}

func listPorts() {
	fmt.Println("MIDI INPUTS:")
	for i, p := range midi.GetInPorts() {
		fmt.Printf("  [%d] %s\n", i, p.String())
	}
	fmt.Println("MIDI OUTPUTS:")
	for i, p := range midi.GetOutPorts() {
		fmt.Printf("  [%d] %s\n", i, p.String())
	}
}

// 数値指定ならインデックスでポートを取得。名前指定なら部分一致で検索する
func findInPort(spec string) (drivers.In, error) {
	if n, err := strconv.Atoi(spec); err == nil {
		ports := midi.GetInPorts()
		if n < 0 || n >= len(ports) {
			return nil, fmt.Errorf("input port index %d out of range (0..%d)", n, len(ports)-1)
		}
		return ports[n], nil
	}
	return midi.FindInPort(spec)
}

func findOutPort(spec string) (drivers.Out, error) {
	if n, err := strconv.Atoi(spec); err == nil {
		ports := midi.GetOutPorts()
		if n < 0 || n >= len(ports) {
			return nil, fmt.Errorf("output port index %d out of range (0..%d)", n, len(ports)-1)
		}
		return ports[n], nil
	}
	return midi.FindOutPort(spec)
}

func run(inSpec, outSpec, script string, verbose bool) error {
	defer midi.CloseDriver()

	in, err := findInPort(inSpec)
	if err != nil {
		return fmt.Errorf("input not found: %w", err)
	}
	out, err := findOutPort(outSpec)
	if err != nil {
		return fmt.Errorf("output not found: %w", err)
	}

	send, err := midi.SendTo(out)
	if err != nil {
		return err
	}

	logger := &eventLogger{enabled: verbose}

	// 送信時にラウンドバッファへ記録
	sendFn := func(msg midi.Message) error {
		logger.recordOut(msg)
		return send(msg)
	}

	L := lua.NewState()
	defer L.Close()
	registerAPI(L, sendFn)

	if err := L.DoFile(script); err != nil {
		return fmt.Errorf("lua load: %w", err)
	}

	// 起動時フック: Lua 側で on_startup() が定義されていれば呼ぶ
	callLua(L, "on_startup")

	var mu sync.Mutex

	stop, err := midi.ListenTo(in, func(msg midi.Message, ts int32) {
		mu.Lock()
		defer mu.Unlock()
		logger.beginRound(msg)
		dispatch(L, msg)
		logger.endRound()
	}, midi.UseSysEx())
	if err != nil {
		return err
	}

	fmt.Printf("Remapping: %s -> %s (script: %s)\n", in.String(), out.String(), script)
	if verbose {
		fmt.Println("Verbose mode: printing every MIDI event")
	}
	fmt.Println("Press Ctrl+C to quit.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	signal.Stop(sig)

	fmt.Println("\nShutting down...")
	stop()
	return nil
}
