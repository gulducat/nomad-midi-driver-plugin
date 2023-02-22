variable "dir" {
  # does this path make you go "hmmmm" ?
  default = "/Users/danielbennett/git/gulducat/nomad-midi-driver-plugin"
}

job "drums" {
  group "g" {
    task "breakbeat" {
      driver = "midi-portmidi"
      config {
        midi_file = "${var.dir}/example/breakbeat.mid"
        #midi_note = "${var.whatever}"
        port_name = "Drums"
        # bars = 8
        # tempo = ??
        # song = "orchestrate me, baby"
        # IF this is absent, you get to effectively disable the lock
        # lock = ??
        # ^ don't neeeed to, since nomad starts only one instance of the plugin binary,
        # but if someone wanted to orchestrate multiple tunes, then they'd need separate locks.
        # ...
        # OK to protect the main driver process, it can fork exec itself to run the actual midi,
        # then it could also keep playing if nomad stops, then recover and pick itself back up...
      }
    }
#    task "break_exec" {
#      driver = "raw_exec"
#      config {
#        command = "/Users/danielbennett/git/gulducat/nomad-midi-driver-plugin/midi-portmidi"
#        args = [
#          "Bus 2",
#          "/Users/danielbennett/git/gulducat/nomad-midi-driver-plugin/example/breakbeat.mid",
#        ]
#      }
#    }
  }
}
