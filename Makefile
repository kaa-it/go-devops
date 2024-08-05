.PHONY: build pg doc swagger

SERVER_VERSION := 1.0.1
AGENT_VERSION := 1.0.0

COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell date +'%Y/%m/%d %H:%M:%S')

build:
	go build -o agent -ldflags \
		"-X github.com/kaa-it/go-devops/internal/buildconfig.buildVersion=${AGENT_VERSION} \
		-X 'github.com/kaa-it/go-devops/internal/buildconfig.buildDate=${DATE}' \
		-X github.com/kaa-it/go-devops/internal/buildconfig.buildCommit=${COMMIT}" ./cmd/agent ;
	go build -o server -ldflags \
		"-X github.com/kaa-it/go-devops/internal/buildconfig.buildVersion=${SERVER_VERSION} \
		-X 'github.com/kaa-it/go-devops/internal/buildconfig.buildDate=${DATE}' \
		-X github.com/kaa-it/go-devops/internal/buildconfig.buildCommit=${COMMIT}" ./cmd/server ;
	go build -o staticlint ./cmd/staticlint/main.go ;

install_tools:
	go install -v golang.org/x/tools/cmd/godoc@latest
	go install github.com/swaggo/swag/cmd/swag@latest

swagger:
	swag init --output ./swagger/ \
    -d ./internal/server/http/rest,./internal/server/http/rest/service,./internal/server/http/rest/viewing,./internal/server/http/rest/updating,./internal/api \
    -g doc.go

pg:
	docker compose up -d ;

doc:
	godoc -http=:9999

run_server:
	CompileDaemon -command='./server -d postgres://ak:postgres@localhost:5432/devops -a :8089 -k xxx' \
	-build="go build -o server ./cmd/server"

run_agent:
	./agent -a "localhost:8089" -k "xxx"

format:
	goimports -w -local github.com/kaa-it/go-devops .

cover:
	go test -v -coverprofile cover.out ./... && grep -v "mock_" cover.out > filtered_cover.out && go tool cover -func=filtered_cover.out

test:
	go vet --vettool=$(which statictest) ./... ;
	metricstest -test.v -test.run=^TestIteration1$$ \
                -binary-path=./server ;
	metricstest -test.v -test.run=^TestIteration2[AB]*$$ \
                -source-path=. \
                -agent-binary-path=./agent ;
	metricstest -test.v -test.run=^TestIteration3[AB]*$$ \
                -source-path=. \
                -agent-binary-path=./agent \
                -binary-path=./server ;
	export SERVER_PORT=9090 && \
    export ADDRESS="localhost:9090" && \
    export TEMP_FILE=test && \
	metricstest -test.v -test.run=^TestIteration4$$ \
		-agent-binary-path=./agent \
		-binary-path=./server \
		-server-port=9090 \
		-source-path=. ;
	export SERVER_PORT=9090 && \
    export ADDRESS="localhost:9090" && \
    export TEMP_FILE=test && \
    metricstest -test.v -test.run=^TestIteration5$$ \
	  -agent-binary-path=./agent \
	  -binary-path=./server \
	  -server-port=9090 \
	  -source-path=. ;
	export SERVER_PORT=9090 && \
    export ADDRESS="localhost:9090" && \
    export TEMP_FILE=test && \
	metricstest -test.v -test.run=^TestIteration6$$ \
                -agent-binary-path=./agent \
                -binary-path=./server \
                -server-port=9090 \
                -source-path=. ;
	export SERVER_PORT=9090 && \
    export ADDRESS="localhost:9090" && \
    export TEMP_FILE=test && \
	metricstest -test.v -test.run=^TestIteration7$$ \
                -agent-binary-path=./agent \
                -binary-path=./server \
                -server-port=9090 \
                -source-path=. ;
	export SERVER_PORT=9090 && \
    export ADDRESS="localhost:9090" && \
    export TEMP_FILE=test && \
	metricstest -test.v -test.run=^TestIteration8$$ \
                -agent-binary-path=./agent \
                -binary-path=./server \
                -server-port=9090 \
                -source-path=. ;
	export SERVER_PORT=9898 && \
	export ADDRESS="localhost:9898" && \
	export TEMP_FILE=/tmp/123.json && \
	metricstest -test.v -test.run=^TestIteration9$$ \
	-agent-binary-path=./agent \
	-binary-path=./server \
	-file-storage-path=/tmp/123.json \
	-server-port=9898 \
	-source-path=. ;
