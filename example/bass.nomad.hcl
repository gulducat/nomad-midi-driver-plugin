job "bass" {
  task "midi" {
    driver = "midi-portmidi"
    config {
      song = "one"
      midi_file = "example/bass.mid"
      port_name = "bass"
      bars = 2
    }
  }
}
