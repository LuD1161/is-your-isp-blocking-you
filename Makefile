APP=is-your-isp-blocking-you
# https://medium.com/the-go-journey/adding-version-information-to-go-binaries-e1b79878f6f2
GIT_COMMIT=$(shell git rev-parse --short=10 HEAD)

.PHONY: build-and-execute
build-and-execute:
	chmod +x ./set-env-vars.sh && . ./set-env-vars.sh
	go build -ldflags "-X main.GitCommit=${GIT_COMMIT}" -o ${APP} main.go && chmod +x ./${APP} && ./${APP}

.PHONY: generate_executables
generate_executables:
	# Generate windows binary
	env GOOS=windows GOARCH=amd64 go build -ldflags "-X main.GitCommit=${GIT_COMMIT}" -o ./binaries/${APP}-win-x64.exe main.go && chmod +x ./${APP} && ./${APP}
	env GOOS=windows GOARCH=386 go build -ldflags "-X main.GitCommit=${GIT_COMMIT}" -o ./binaries/${APP}-win-x86.exe main.go && chmod +x ./${APP} && ./${APP}
	# Generate linux binary
	env GOOS=linux GOARCH=amd64 go build -ldflags "-X main.GitCommit=${GIT_COMMIT}" -o ./binaries/${APP}-linux-x64 main.go && chmod +x ./${APP} && ./${APP}
	env GOOS=linux GOARCH=386 go build -ldflags "-X main.GitCommit=${GIT_COMMIT}" -o ./binaries/${APP}-linux-x86 main.go && chmod +x ./${APP} && ./${APP}
	# Generate darwin binary
	env GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.GitCommit=${GIT_COMMIT}" -o ./binaries/${APP}-mac-x64 main.go && chmod +x ./${APP} && ./${APP}
	env GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.GitCommit=${GIT_COMMIT}" -o ./binaries/${APP}-mac-m1 main.go && chmod +x ./${APP} && ./${APP}

.PHONY: build
build:
	go build -ldflags "-X main.GitCommit=${GIT_COMMIT}" -o ${APP} main.go

.PHONY: run
run:
	go run main.go

.PHONY: debug
debug: 
	export DEBUG=True && make build-and-execute
	
.PHONY: prod
prod: 
	export PROD=True && make build-and-execute