job "[[ .midi.part ]]" {
  group "g" {
    task "part" {
      driver = "midi-portmidi"
      config {
        song      = "[[ .midi.song ]]"
        midi_file = "[[ .midi.file_dir ]]/[[ .midi.part ]].mid"
        port_name = "[[ .midi.part ]]"
        bars      = [[ index .midi.bars .midi.part ]]
      }
      env {
        LOG_LEVEL = "DEBUG"
      }
    }
  }
}
