watch() {
  local msg
  while true; do
    msg='waiting for nomad...'
    curl -s localhost:4646 | grep -q . \
      && msg="$(nomad status)"
    clear
    echo "$msg"
    printf "\n"
    sleep 1
  done
}

_exec() {
  echo "++ $*"
  eval "$*"
}

_run (){
  _exec nomad-pack run example/pack --var=midi.part=$part
}

run() {
  for part in $*; do _run $part; done
}

_stop() {
  local job="$1"
  shift
  if [ $job = all ]; then
    echo 'stopping all jobs'
    nomad status | awk '/running|pending|dead/ {print$1}' \
      | while read -r j; do nomad stop -detach $* $j; done
    return
  fi
  _exec nomad stop -detach $* $job
}

stop() {
  for part in $*; do _stop $part; done
}
