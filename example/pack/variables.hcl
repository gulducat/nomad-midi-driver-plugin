variable "song" {
  description = "a clock will coordinate all parts in a song"
  type = string
  default = "orchestrate me, baby"
}

variable "part" {
  description = "{part}.mid + a midi port named {part}"
  type = string
}

# weird but ok
variable "bars" {
  description = "map of number of bars per part"
  type = map(string)
  default = {
    "mallet": 1,
    "drums": 2,
    "brass": 4,
    "strings": 8,
    "arp": 2,
    "bass": 2,
    "hats": 1,
  }
}

variable "file_dir" {
  description = "location of .mid files"
  type = string
  default = "example"
  #default = "/Users/danielbennett/git/gulducat/nomad-midi-driver-plugin/example"
}
