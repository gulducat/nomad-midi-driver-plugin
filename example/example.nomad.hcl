job "song" {
  group "g" {
    task "mallet" {
      driver = "midi-portmidi"
      config {
        # the "song" is used to synchronize all the tasks with a single clock
        song = "one"
        # path to midi file, relative to the location of the driver binary
        # i.e. relative to Nomad's plugin_dir
        midi_file = "example/mallet.mid"
        # the name of the virtual MIDI port
        port_name = "mallet"
        # how long the midi file is
        bars = 1
      }
    }
    task "drums" {
      driver = "midi-portmidi"
      config {
        song = "one"
        midi_file = "example/drums.mid"
        port_name = "drums"
        bars = 2
      }
    }
    task "brass" {
      driver = "midi-portmidi"
      config {
        song = "one"
        midi_file = "example/brass.mid"
        port_name = "brass"
        bars = 4
      }
    }
    task "strings" {
      driver = "midi-portmidi"
      config {
        song = "one"
        midi_file = "example/strings.mid"
        port_name = "strings"
        bars = 8
      }
    }
    task "bass" {
      driver = "midi-portmidi"
      config {
        song = "one"
        midi_file = "example/bass.mid"
        port_name = "bass"
        bars = 2
      }
    }
    task "arp" {
      driver = "midi-portmidi"
      config {
        song = "one"
        midi_file = "example/arp.mid"
        port_name = "arp"
        bars = 2
      }
    }
    task "cat" {
      driver = "midi-portmidi"
      config {
        song = "one"
        midi_file = "example/cat.mid"
        port_name = "cat"
        bars = 8
      }
    }
  }
}
