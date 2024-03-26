build:
	go build -o agent ./cmd/agent

incr1:
	go vet --vettool=$(which statictest) ./...
	devopstest -test.v -test.run=^TestIteration1$$ -agent-binary-path=./agent