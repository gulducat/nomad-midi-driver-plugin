locals {
  dir = "/Users/danielbennett/git/gulducat/nomad-midi-driver-plugin/example"
}

# this is to test timing -- they should all hit at the same time

job "hit" {
  group "g" {
    task "hit1" {
      driver = "midi-portmidi"
      config {
        song      = "one"
        midi_file = "${local.dir}/hit.mid"
        port_name = "hit"
        bars = 1
      }
    }
    task "hit2" {
      driver = "midi-portmidi"
      config {
        song      = "one"
        midi_file = "${local.dir}/hit.mid"
        port_name = "hit2"
        bars = 1
      }
    }
    task "hit3" {
      driver = "midi-portmidi"
      config {
        song      = "one"
        midi_file = "${local.dir}/hit.mid"
        port_name = "hit3"
        bars = 1
      }
    }
  }
}
