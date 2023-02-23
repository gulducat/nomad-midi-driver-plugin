package main

import (
	"context"
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
	logger := hclog.Default()
	var players []*nomidi.Player
	ctx, stop := context.WithTimeout(context.Background(), time.Second*30)
	defer stop()

	clock := nomidi.GetClock("cli")
	go clock.Tick(ctx)

	for f, bars := range map[string]int{
		"mallet":  1,
		"drums":   2,
		"brass":   2,
		"strings": 8,
		"arp":     2,
		"bass":    2,
	} {
		cfg := nomidi.TaskConfig{
			PortName: f,
			MidiFile: "example/" + f + ".mid",
			Bars:     bars,
		}
		p := nomidi.NewPlayer(logger, cfg)
		clock.Subscribe(p)
		players = append(players, p)
		go p.Play(ctx)
	}

	for _, p := range players {
		p.Wait(ctx)
	}
}

func cli(port, file string) {
	// TODO: handle signals?
	ctx := context.Background()

	clock := nomidi.NewClock("cli")
	defer nomidi.DeleteClock("cli")
	go clock.Tick(ctx)

	logger := hclog.Default()
	cfg := nomidi.TaskConfig{
		PortName: port,
		MidiFile: file,
	}
	player := nomidi.NewPlayer(logger, cfg)
	clock.Subscribe(player)
	defer clock.Unsubscribe(player)

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
