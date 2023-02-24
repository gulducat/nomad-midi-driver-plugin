job "drums" {
  task "midi" {
    driver = "midi-portmidi"
    config {
      #song = "one"
      midi_file = "example/drums.mid"
      port_name = "drums"
      bars = 2
    }
  }
}
