locals {
  dir = "/Users/danielbennett/git/gulducat/nomad-midi-driver-plugin/example"
}

job "hit" {
  group "g" {
    task "hit1" {
      driver = "midi-portmidi"
      config {
        song      = "one"
        midi_file = "${local.dir}/hit-bar.mid"
        port_name = "hit"
        bars = 1
      }
    }
#    task "hit2" {
#      driver = "midi-portmidi"
#      config {
#        song      = "one"
#        midi_file = "${local.dir}/hit-bar.mid"
#        port_name = "hit2"
#        bars = 1
#      }
#    }
#    task "hit3" {
#      driver = "midi-portmidi"
#      config {
#        song      = "one"
#        midi_file = "${local.dir}/hit-bar.mid"
#        port_name = "hit3"
#        bars = 1
#      }
#    }
  }
}
