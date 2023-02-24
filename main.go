package main

// TODO: delete all the stuff I don't need

import (
	"context"
	"github.com/gulducat/nomad-midi-driver-plugin/maestro"
	"gitlab.com/gomidi/midi/v2"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins"

	_ "gitlab.com/gomidi/midi/v2/drivers/portmididrv" // autoregisters driver
)

var parts = map[string]int{
	// part :  # of bars
	"mallet":  1,
	"drums":   2,
	"brass":   4,
	"strings": 8,
	"arp":     2,
	"bass":    2,
	"hats":    1,
	"cat":     8,
}

func main() {
	defer midi.CloseDriver()

	// allow direct CLI usage
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()
	if len(os.Args) == 2 && os.Args[1] == "test" {
		testSync(ctx)
		return
	}
	if len(os.Args) == 4 {
		cli(ctx, os.Args[1], os.Args[2], os.Args[3])
		return
	}

	// or, be a nomad plugin
	plugins.Serve(factory)
}

func testSync(ctx context.Context) {
	/* testing concurrent files, make sure they are sync'd */
	logger := hclog.Default()
	var players []*maestro.Player

	clock := maestro.GetClock("cli")
	go clock.Tick(ctx)

	for f, bars := range parts {
		cfg := maestro.TaskConfig{
			PortName: f,
			MidiFile: "example/" + f + ".mid",
			Bars:     bars,
		}
		p := maestro.NewPlayer(logger, cfg)
		clock.Subscribe(p)
		players = append(players, p)
		go p.Play(ctx)
	}

	ctx, stop := context.WithTimeout(context.Background(), time.Minute)
	defer stop()
	for _, p := range players {
		p.Wait(ctx)
	}
}

func cli(ctx context.Context, port, file, bars string) {
	clock := maestro.GetClock("cli")
	defer maestro.DeleteClock("cli")
	go clock.Tick(ctx)

	logger := hclog.Default()
	logger.SetLevel(hclog.Debug)
	numBars, err := strconv.Atoi(bars)
	if err != nil {
		log.Fatal(err)
	}
	cfg := maestro.TaskConfig{
		PortName: port,
		MidiFile: file,
		Bars:     numBars,
	}
	player := maestro.NewPlayer(logger, cfg)
	clock.Subscribe(player)
	defer clock.Unsubscribe(player)

	go player.Play(ctx)
	err = player.Wait(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

// factory returns a new instance of a nomad driver plugin
func factory(log hclog.Logger) interface{} {
	return maestro.NewPlugin(log)
}
