PLUGIN_BINARY=midi-portmidi
export GO111MODULE=on

default: nomad

kill:
	pkill nomad || true
	ps aux | grep -E 'nomad|midi' | grep -v grep || true

nomad: kill build
	nomad agent -dev -config=./example/agent.hcl | grep -i midi &

wait:
	while true; do nomad status && break ; sleep 1; done

.PHONY: example
example: wait
	nomad run example/example.nomad.hcl

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf ${PLUGIN_BINARY}

build:
	go build -o ${PLUGIN_BINARY} .

test: build
	./$(PLUGIN_BINARY) test
