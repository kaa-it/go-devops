build:
	go build -o agent ./cmd/agent

incr1:
	devopstest -test.v -test.run=^TestIteration1$$ -agent-binary-path=./agent