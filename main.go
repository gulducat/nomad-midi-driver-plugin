package main

import (
	"context"
	"log"
	"os"

	"github.com/gulducat/nomad-midi-driver-plugin/nomidi"
	"gitlab.com/gomidi/midi/v2"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins"

	_ "gitlab.com/gomidi/midi/v2/drivers/portmididrv" // autoregisters driver
)

func main() {
	defer midi.CloseDriver()
	// allow direct CLI usage
	if len(os.Args) == 3 {
		cli(os.Args[1], os.Args[2])
		return
	}
	// or, be a nomad plugin
	plugins.Serve(factory)
}

func cli(port, midiFile string) {
	// TODO: handle signals?
	ctx := context.Background()

	logger := hclog.Default()
	player := nomidi.NewPlayer(logger)
	go player.Play(ctx, port, midiFile)
	err := player.Wait(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

// factory returns a new instance of a nomad driver plugin
func factory(log hclog.Logger) interface{} {
	return nomidi.NewPlugin(log)
}
