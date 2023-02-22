locals {
  dir = "/Users/danielbennett/git/gulducat/nomad-midi-driver-plugin/example"
}
job "e-piano" {
  task "midi" {
    driver = "midi-portmidi"
    config {
      song = "one"
      midi_file = "${local.dir}/e-piano.mid"
      port_name = "e-piano"
      bars = 4
    }
  }
}
