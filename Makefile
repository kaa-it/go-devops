build:
	go build -o agent ./cmd/agent
	go build -o server ./cmd/server

test:
	go vet --vettool=$(which statictest) ./...
	devopstest -test.v -test.run=^TestIteration1$$ -agent-binary-path=./agent
	devopstest -test.v -test.run=^TestIteration2[b]*$$ \
                -source-path=. \
                -binary-path=./server
	devopstest -test.v -test.run=^TestIteration3[b]*$ \
                -source-path=. \
                -binary-path=./server