locals {
  dir = "/Users/danielbennett/git/gulducat/nomad-midi-driver-plugin/example"
}
job "bass" {
  task "midi" {
    driver = "midi-portmidi"
    config {
      song = "one"
      midi_file = "${local.dir}/bass.mid"
      port_name = "bass"
      bars = 2
    }
  }
}
