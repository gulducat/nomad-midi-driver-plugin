_exec() {
  echo "$@"
  eval "$@"
}

run (){
  _exec nomad-pack run example/pack --var=midi.part=$1
}

stop() {
  local job="$1"
  shift
  if [ $job = all ]; then
    echo 'stopping all jobs'
    nomad status | awk '/service/ {print$1}' | while read -r j; do nomad stop -detach $* $j; done
    return
  fi
  _exec nomad stop -detach $* $job
}