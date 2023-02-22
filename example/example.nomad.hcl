variable "dir" {
  # does this path make you go "hmmmm" ?
  default = "/Users/danielbennett/git/gulducat/nomad-midi-driver-plugin/example"
}

job "song" {
  group "g" {
    task "hit" {
      driver = "midi-portmidi"
      config {
        song = "one"
        midi_file = "${var.dir}/hit.mid"
        port_name = "hit"
        bars = 1
        # midi_note = "${var.whatever}"
        # tempo = ??
        # song = "orchestrate me, baby"
        # IF this is absent, you get to effectively disable the lock
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
    task "e-piano" {
      driver = "midi-portmidi"
      config {
        song = "one"
        midi_file = "${var.dir}/e-piano.mid"
        port_name = "e-piano"
        bars = 4
      }
    }
    task "drums" {
      driver = "midi-portmidi"
      config {
        song = "one"
        midi_file = "${var.dir}/drums.mid"
        port_name = "drums"
        bars = 8
      }
    }
  }
}
