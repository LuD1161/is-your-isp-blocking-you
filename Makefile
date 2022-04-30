APP=isp-block-checker
# https://medium.com/the-go-journey/adding-version-information-to-go-binaries-e1b79878f6f2
GIT_COMMIT=$(shell git rev-parse --short=10 HEAD)

.PHONY: build-and-execute
build-and-execute:
	chmod +x ./set-env-vars.sh && . ./set-env-vars.sh && go build -ldflags "-X main.GitCommit=${GIT_COMMIT}" -o ${APP} cmd/checker/main.go && chmod +x ./${APP} && ./${APP}

.PHONY: build
build:
	go build -o ${APP} cmd/checker/main.go

.PHONY: run
run:
	go run cmd/checker/main.go

.PHONY: debug
debug: 
	export DEBUG=True && make build-and-execute
	
.PHONY: prod
prod: 
	export PROD=True && make build-and-execute
