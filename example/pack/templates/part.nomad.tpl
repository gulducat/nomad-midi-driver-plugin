job "[[ .midi.part ]]" {
  group "[[ .midi.part ]]" {
    task "[[ .midi.part ]]" {
      driver = "midi-portmidi"
      config {
        song      = "[[ .midi.song ]]"
        port_name = "[[ .midi.part ]]"
        midi_file = "[[ .midi.file_dir ]]/[[ .midi.part ]].mid"
        bars      = [[ index .midi.bars .midi.part ]]
      }
    }
  }
}
