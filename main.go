package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
	lua "github.com/yuin/gopher-lua"
)

func main() {
	inPort := flag.String("in", "", "MIDI input port name (substring match)")
	outPort := flag.String("out", "", "MIDI output port name (substring match)")
	script := flag.String("script", "config.lua", "Lua remap script")
	list := flag.Bool("list", false, "list MIDI ports and exit")
	flag.Parse()

	if *list {
		fmt.Println("INS:")
		for _, p := range midi.GetInPorts() {
			fmt.Printf("  [%d] %s\n", p.Number(), p.String())
		}
		fmt.Println("OUTS:")
		for _, p := range midi.GetOutPorts() {
			fmt.Printf("  [%d] %s\n", p.Number(), p.String())
		}
		return
	}

	in, err := midi.FindInPort(*inPort)
	if err != nil {
		log.Fatalf("input not found: %v", err)
	}
	out, err := midi.FindOutPort(*outPort)
	if err != nil {
		log.Fatalf("output not found: %v", err)
	}

	send, err := midi.SendTo(out)
	if err != nil {
		log.Fatal(err)
	}

	L := lua.NewState()
	defer L.Close()
	registerAPI(L, send)
	if err := L.DoFile(*script); err != nil {
		log.Fatalf("lua load: %v", err)
	}

	// Lua VM はシングルスレッドなので mutex で保護
	var mu sync.Mutex

	stop, err := midi.ListenTo(in, func(msg midi.Message, ts int32) {
		mu.Lock()
		defer mu.Unlock()
		dispatch(L, msg)
	}, midi.UseSysEx())
	if err != nil {
		log.Fatal(err)
	}
	defer stop()

	fmt.Printf("Remapping: %s -> %s (script: %s)\n", in.String(), out.String(), *script)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	midi.CloseDriver()
}
