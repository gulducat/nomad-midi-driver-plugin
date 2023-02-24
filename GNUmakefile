PLUGIN_BINARY=midi-portmidi
export GO111MODULE=on

default: nomad

nomad: kill build
	nomad agent -dev -config=./example/agent.hcl | grep -i midi &

reset:
	nomad status | awk '/service/ {print$$1}' | xargs -P1 nomad stop -purge -detach
	pkill midi-portmidi

kill:
	pkill nomad midi-portmidi || true
	ps aux | grep -E 'nomad|midi' | grep -v grep || true

wait:
	while true; do nomad status && break ; sleep 1; done

example: wait
	nomad run example/example.nomad.hcl

clean: ## Remove build artifacts
	rm -rf ${PLUGIN_BINARY}

build:
	go build -o ${PLUGIN_BINARY} .

test: build
	./$(PLUGIN_BINARY) test

.PHONY: default reset kill nomad wait example clean build test
