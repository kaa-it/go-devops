build:
	go build -o agent ./cmd/agent
	go build -o server ./cmd/server

test:
	go vet --vettool=$(which statictest) ./...
	metricstest -test.v -test.run=^TestIteration1$$ \
                -binary-path=./server
	metricstest -test.v -test.run=^TestIteration2[AB]*$$ \
                -source-path=. \
                -agent-binary-path=./agent
	metricstest -test.v -test.run=^TestIteration3[AB]*$$ \
                -source-path=. \
                -agent-binary-path=./agent \
                -binary-path=./server
	export SERVER_PORT=9090 && \
    export ADDRESS="localhost:9090" && \
    export TEMP_FILE=test && \
	metricstest -test.v -test.run=^TestIteration4$ \
		-agent-binary-path=./agent \
		-binary-path=./server \
		-server-port=9090 \
		-source-path=.