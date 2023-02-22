locals {
  dir = "/Users/danielbennett/git/gulducat/nomad-midi-driver-plugin/example"
}
job "drums" {
  task "midi" {
    driver = "midi-portmidi"
    config {
      song = "one"
      midi_file = "${local.dir}/drums.mid"
      port_name = "drums"
      bars = 8
    }
  }
}
