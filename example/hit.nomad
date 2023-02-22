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
        port_name = "Hit1"
      }
    }
#    task "hit2" {
#      driver = "midi-portmidi"
#      config {
#        song      = "one"
#        midi_file = "${local.dir}/hit-bar.mid"
#        port_name = "Hit2"
#      }
#    }
#    task "hit3" {
#      driver = "midi-portmidi"
#      config {
#        song      = "one"
#        midi_file = "${local.dir}/hit-bar.mid"
#        port_name = "Hit3"
#      }
#    }
  }
}
