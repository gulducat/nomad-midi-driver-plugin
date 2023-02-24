Maestro
==========

Nomad is actually an orchestrator.

Maestro is A MIDI
[task driver](https://developer.hashicorp.com/nomad/docs/drivers)
for HashiCorp Nomad that sends MIDI signals from .mid files
to music software like Ableton Live.

Example demo video for our Nomad team Hack Week available
[here](https://drive.google.com/file/d/1TAL5d-UpkvrS_IvQNHyDDgLZJGdAfAOS/view?usp=sharing)!

Requirements
-------------------

- [Go](https://golang.org/doc/install) v1.18 or later (to compile the plugin)
- [Nomad](https://www.nomadproject.io/downloads.html) v0.9+ (to run the plugin)
- MacOS (maybe it works on windows, idk)
- [portmidi](https://github.com/PortMidi/portmidi) C lib to play the MIDI
  (`brew install portmidi`)
- [Virtual MIDI ports](https://help.ableton.com/hc/en-us/articles/209774225-Setting-up-a-virtual-MIDI-bus)
- Some music software (e.g. Ableton Live)
- [gomidi/midi](https://github.com/gomidi/midi) has some neat tools available too.

Building the Plugin
-------------------

```sh
$ make build
```

## Deploying Driver Plugins in Nomad

Start Nomad with the example agent config and set the plugin dir to where the binary is.

`make build` puts it in the current directory.

```sh
$ nomad agent -dev -config=./example/agent.hcl -plugin-dir=$(pwd)
```

## Configuring your machine

In my setup, each musical "part" needs these things:

- MIDI file
- [virtual MIDI port](https://help.ableton.com/hc/en-us/articles/209774225-Setting-up-a-virtual-MIDI-bus)
- track in your music software DAW that listens on that port
  (and be sure the track is armed)
- task in a job with the `"midi-portmidi"` driver (see `example/example.nomad.hcl`)
  that connects all these things together.

I export my `.mid` files from Ableton, which only does single-channel files, hence all the separate ports.
You can generate them differently and take advantage of channels on one port, if you can work that out.

## Running tasks

Have a look in `example/` for various examples or whatever.

`source example/funcy.sh` for some helpful functions to `run` or `stop` different parts.

I'm pretty tired at the end of Hack Week, so I'm just gonna leave this here.

Good luck!
