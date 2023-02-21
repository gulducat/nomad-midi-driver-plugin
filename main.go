package main

import (
	"github.com/gulducat/nomad-midi-driver-plugin/nomidi"
	"log"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins"
)

func main() {
	// allow direct CLI usage
	if len(os.Args) == 3 {
		// TODO: handle signals
		err := nomidi.Play(os.Args[1], os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	plugins.Serve(factory)
}

// factory returns a new instance of a nomad driver plugin
func factory(log hclog.Logger) interface{} {
	return nomidi.NewPlugin(log)
}
