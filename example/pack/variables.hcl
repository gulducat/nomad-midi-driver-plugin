variable "song" {
  type = string
  description = "a clock will coordinate all parts in the same song"
  default = "orchestrate me, baby"
}

variable "part" {
  type = string
  description = "examples/{part}.mid and a midi port named {part}"
}

# weird but ok
variable "bars" {
  type = map(string)
  default = {
    "mallet": 1,
    "drums": 2,
    "brass": 2,
    "strings": 8,
    "arp": 2,
    "bass": 2,
  }
}

variable "file_dir" {
  type = string
  description = "location of .mid files"
  default = "/Users/danielbennett/git/gulducat/nomad-midi-driver-plugin/example"
}
