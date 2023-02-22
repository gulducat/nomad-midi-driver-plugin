package main

import (
	"context"
	"fmt"
	"github.com/gulducat/nomad-midi-driver-plugin/nomidi"
	"gitlab.com/gomidi/midi/v2"
	"log"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins"

	_ "gitlab.com/gomidi/midi/v2/drivers/portmididrv" // autoregisters driver
)

func main() {
	defer midi.CloseDriver()

	//testSync()
	//return

	// allow direct CLI usage
	if len(os.Args) == 3 {
		cli(os.Args[1], os.Args[2])
		return
	}
	// or, be a nomad plugin
	plugins.Serve(factory)
}

func testSync() {
	/* testing concurrent files, make sure they are sync'd */
	f := "example/hit.mid"
	logger := hclog.Default()
	var players []*nomidi.Player
	ctx, stop := context.WithTimeout(context.Background(), time.Second*8)
	defer stop()

	clock := nomidi.GetClock("cli")
	go clock.Tick(ctx)

	for i := 1; i <= 4; i++ {
		port := fmt.Sprintf("Hit%d", i)
		p := nomidi.NewPlayer(logger, port, f)
		clock.Subscribe(p)
		players = append(players, p)
		go p.Play(ctx)
	}

	for _, p := range players {
		p.Wait(ctx)
		clock.Unsubscribe(p)
	}
	if err := nomidi.DeleteClock("cli"); err != nil {
		log.Fatal(err)
	}
}

func cli(port, midiFile string) {
	// TODO: handle signals?
	ctx := context.Background()

	logger := hclog.Default()
	player := nomidi.NewPlayer(logger, port, midiFile)

	clock := nomidi.NewClock("cli")
	defer nomidi.DeleteClock("cli")
	clock.Subscribe(player)
	defer clock.Unsubscribe(player)

	go clock.Tick(ctx)
	go player.Play(ctx)

	err := player.Wait(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

// factory returns a new instance of a nomad driver plugin
func factory(log hclog.Logger) interface{} {
	return nomidi.NewPlugin(log)
}
