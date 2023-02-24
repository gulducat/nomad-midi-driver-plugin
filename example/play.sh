#!/bin/bash
source example/funcy.sh

_sleep() {
  printf '\n'
  for x in $(seq 1 $1); do
    printf '.'
    sleep 1
  done
  printf '\n'
}

 run mallet
_sleep 8
 run drums
_sleep 7
 run brass strings
_sleep 32
 run bass
_sleep 16
 run arp
 stop brass
_sleep 15 # 16 for overlap?
 run hats cat
 stop strings
_sleep 28
 run brass strings
 stop bass arp hats cat
_sleep 16
 stop drums
_sleep 8
 stop mallet
_sleep 4
 stop brass strings
