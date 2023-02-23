variable "dir" {
  # does this path make you go "hmmmm" ?
  default = "/Users/danielbennett/git/gulducat/nomad-midi-driver-plugin/example"
}

job "song" {
  group "g" {
    task "mallet" {
      driver = "midi-portmidi"
      config {
        song = "one"
        midi_file = "${var.dir}/mallet.mid"
        port_name = "mallet"
        bars = 1
        # midi_note = "${var.whatever}"
        # tempo = ??
        # song = "orchestrate me, baby"
        # IF this is absent, you get to effectively disable the lock
      }
    }
    task "drums" {
      driver = "midi-portmidi"
      config {
        song = "one"
        midi_file = "${var.dir}/drums.mid"
        port_name = "drums"
        bars = 2
      }
    }
    task "brass" {
      driver = "midi-portmidi"
      config {
        song = "one"
        midi_file = "${var.dir}/brass.mid"
        port_name = "brass"
        bars = 2
      }
    }
    task "strings" {
      driver = "midi-portmidi"
      config {
        song = "one"
        midi_file = "${var.dir}/strings.mid"
        port_name = "strings"
        bars = 8
      }
    }
    task "arp" {
      driver = "midi-portmidi"
      config {
        song = "one"
        midi_file = "${var.dir}/arp.mid"
        port_name = "arp"
        bars = 2
      }
    }
    task "bass" {
      driver = "midi-portmidi"
      config {
        song = "one"
        midi_file = "${var.dir}/bass.mid"
        port_name = "bass"
        bars = 2
      }
    }
  }
}
