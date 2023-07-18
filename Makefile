ifndef GOBIN
GOBIN := $(GOPATH)/bin
endif

ifdef JANUS_PORT
JANUS_PORT := $(JANUS_PORT)
else
JANUS_PORT := 24889
endif

ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
JANUS_DIR := "/go/src/github.com/denuoweb/janus"
GO_VERSION := "1.18"
ALPINE_VERSION := "3.16"
DOCKER_ACCOUNT := ripply


# Latest commit hash
GIT_SHA=$(shell git rev-parse HEAD)

# If working copy has changes, append `-local` to hash
GIT_DIFF=$(shell git diff -s --exit-code || echo "-local")
GIT_REV=$(GIT_SHA)$(GIT_DIFF)
GIT_TAG=$(shell git describe --tags 2>/dev/null)

ifeq ($(GIT_TAG),)
GIT_TAG := $(GIT_REV)
else
GIT_TAG := $(GIT_TAG)$(GIT_DIFF)
endif

check-env:
ifndef GOPATH
	$(error GOPATH is undefined)
endif

.PHONY: install
install:
	go install \
		-ldflags "-X 'github.com/denuoweb/janus/pkg/params.GitSha=`./sha.sh``git diff -s --exit-code || echo \"-local\"`'" \
		github.com/denuoweb/janus

.PHONY: release
release: darwin linux windows

.PHONY: darwin
darwin: build-darwin-amd64 tar-gz-darwin-amd64 build-darwin-arm64 tar-gz-darwin-arm64

.PHONY: linux
linux: build-linux-386 tar-gz-linux-386 build-linux-amd64 tar-gz-linux-amd64 build-linux-arm tar-gz-linux-arm build-linux-arm64 tar-gz-linux-arm64 build-linux-ppc64 tar-gz-linux-ppc64 build-linux-ppc64le tar-gz-linux-ppc64le build-linux-mips tar-gz-linux-mips build-linux-mipsle tar-gz-linux-mipsle build-linux-riscv64 tar-gz-linux-riscv64 build-linux-s390x tar-gz-linux-s390x

.PHONY: windows
windows: build-windows-386 tar-gz-windows-386 build-windows-amd64 tar-gz-windows-amd64 build-windows-arm64 tar-gz-windows-arm64
	echo hey
#	GOOS=linux GOARCH=arm64 go build -o ./build/janus-linux-arm64 github.com/denuoweb/janus/cli/janus

docker-build-go-build:
	docker build -t htmlcoin/go-build.janus -f ./docker/go-build.Dockerfile --build-arg GO_VERSION=$(GO_VERSION) .

tar-gz-%:
	mv $(ROOT_DIR)/build/bin/janus-$(shell echo $@ | sed s/tar-gz-// | sed 's/-/\n/' | awk 'NR==1')-$(shell echo $@ | sed s/tar-gz-// | sed 's/-/\n/' | awk 'NR==2') $(ROOT_DIR)/build/bin/janus
	tar -czf $(ROOT_DIR)/build/janus-$(GIT_TAG)-$(shell echo $@ | sed s/tar-gz-// | sed 's/-/\n/' | awk 'NR==1' | sed s/darwin/osx/)-$(shell echo $@ | sed s/tar-gz-// | sed 's/-/\n/' | awk 'NR==2').tar.gz $(ROOT_DIR)/build/bin/janus
	mv $(ROOT_DIR)/build/bin/janus $(ROOT_DIR)/build/bin/janus-$(shell echo $@ | sed s/tar-gz-// | sed 's/-/\n/' | awk 'NR==1')-$(shell echo $@ | sed s/tar-gz-// | sed 's/-/\n/' | awk 'NR==2')

# build-os-arch
build-%: docker-build-go-build
	docker run \
		--privileged \
		--rm \
		-v `pwd`/build:/build \
		-v `pwd`:$(JANUS_DIR) \
		-w $(JANUS_DIR) \
		-e GOOS=$(shell echo $@ | sed s/build-// | sed 's/-/\n/' | awk 'NR==1') \
		-e GOARCH=$(shell echo $@ | sed s/build-// | sed 's/-/\n/' | awk 'NR==2') \
		htmlcoin/go-build.janus \
			build \
			-buildvcs=false \
			-ldflags \
				"-X 'github.com/denuoweb/janus/pkg/params.GitSha=`./sha.sh`'" \
			-o /build/bin/janus-$(shell echo $@ | sed s/build-// | sed 's/-/\n/' | awk 'NR==1')-$(shell echo $@ | sed s/build-// | sed 's/-/\n/' | awk 'NR==2') $(JANUS_DIR)

.PHONY: quick-start
quick-start-regtest:
	cd docker && ./spin_up.regtest.sh && cd ..

.PHONY: quick-start-testnet
quick-start-testnet:
	cd docker && ./spin_up.testnet.sh && cd ..

.PHONY: quick-start-mainnet
quick-start-mainnet:
	cd docker && ./spin_up.mainnet.sh && cd ..

# docker build -t htmlcoin/janus:latest -t htmlcoin/janus:dev -t htmlcoin/janus:${GIT_TAG} -t htmlcoin/janus:${GIT_REV} --build-arg BUILDPLATFORM="$(BUILDPLATFORM)" .
.PHONY: docker-dev
docker-dev:
	docker build -t htmlcoin/janus:latest -t htmlcoin/janus:dev -t htmlcoin/janus:${GIT_TAG} -t htmlcoin/janus:${GIT_REV} --build-arg GO_VERSION=1.18 .

.PHONY: local-dev
local-dev: check-env install
	docker run --rm --name htmlcoin_testchain -d -p 4889:4889 htmlcoin/htmlcoin htmlcoind -regtest -rpcbind=0.0.0.0:4889 -rpcallowip=0.0.0.0/0 -logevents=1 -rpcuser=htmlcoin -rpcpassword=testpasswd -deprecatedrpc=accounts -printtoconsole | true
	sleep 3
	docker cp ${GOPATH}/src/github.com/denuoweb/janus/docker/fill_user_account.sh htmlcoin_testchain:.
	docker exec htmlcoin_testchain /bin/sh -c ./fill_user_account.sh
	HTMLCOIN_RPC=http://htmlcoin:testpasswd@localhost:4889 HTMLCOIN_NETWORK=auto $(GOBIN)/janus --port $(JANUS_PORT) --accounts ./docker/standalone/myaccounts.txt --dev

.PHONY: local-dev-https
local-dev-https: check-env install
	docker run --rm --name htmlcoin_testchain -d -p 4889:4889 htmlcoin/htmlcoin htmlcoind -regtest -rpcbind=0.0.0.0:4889 -rpcallowip=0.0.0.0/0 -logevents=1 -rpcuser=htmlcoin -rpcpassword=testpasswd -deprecatedrpc=accounts -printtoconsole | true
	sleep 3
	docker cp ${GOPATH}/src/github.com/denuoweb/janus/docker/fill_user_account.sh htmlcoin_testchain:.
	docker exec htmlcoin_testchain /bin/sh -c ./fill_user_account.sh > /dev/null&
	HTMLCOIN_RPC=http://htmlcoin:testpasswd@localhost:4889 HTMLCOIN_NETWORK=auto $(GOBIN)/janus --port $(JANUS_PORT) --accounts ./docker/standalone/myaccounts.txt --dev --https-key https/key.pem --https-cert https/cert.pem

.PHONY: local-dev-logs
local-dev-logs: check-env install
	docker run --rm --name htmlcoin_testchain -d -p 4889:4889 htmlcoin/htmlcoin:dev htmlcoind -regtest -rpcbind=0.0.0.0:4889 -rpcallowip=0.0.0.0/0 -logevents=1 -rpcuser=htmlcoin -rpcpassword=testpasswd -deprecatedrpc=accounts -printtoconsole | true
	sleep 3
	docker cp ${GOPATH}/src/github.com/denuoweb/janus/docker/fill_user_account.sh htmlcoin_testchain:.
	docker exec htmlcoin_testchain /bin/sh -c ./fill_user_account.sh
	HTMLCOIN_RPC=http://htmlcoin:testpasswd@localhost:4889 HTMLCOIN_NETWORK=auto $(GOBIN)/janus --port $(JANUS_PORT) --accounts ./docker/standalone/myaccounts.txt --dev > janus_dev_logs.txt

.PHONY: unit-tests
unit-tests: check-env
	go test -v ./... -timeout 50s

docker-build-unit-tests:
	docker build -t htmlcoin/tests.janus -f ./docker/unittests.Dockerfile --build-arg GO_VERSION=$(GO_VERSION) .

docker-unit-tests:
	docker run --rm -v `pwd`:/go/src/github.com/denuoweb/janus htmlcoin/tests.janus

docker-tests: docker-build-unit-tests docker-unit-tests openzeppelin-docker-compose

docker-configure-https: docker-configure-https-build
	docker/setup_self_signed_https.sh

docker-configure-https-build:
	docker build -t htmlcoin/openssl.janus -f ./docker/openssl.Dockerfile ./docker

# -------------------------------------------------------------------------------------------------------------------
# NOTE:
# 	The following make rules are only for local test purposes
#
# 	Both run-janus and run-htmlcoin must be invoked. Invocation order may be independent, 
# 	however it's much simpler to do in the following order:
# 		(1) make run-htmlcoin
# 			To stop htmlcoin node you should invoke: make stop-htmlcoin
# 		(2) make run-janus
# 			To stop Janus service just press Ctrl + C in the running terminal

# Runs current Janus implementation
run-janus:
	@ printf "\nRunning Janus...\n\n"

	go run `pwd`/main.go \
		--htmlcoin-rpc=http://${test_user}:${test_user_passwd}@0.0.0.0:4889 \
		--htmlcoin-network=auto \
		--bind=0.0.0.0 \
		--port=24889 \
		--accounts=`pwd`/docker/standalone/myaccounts.txt \
		--log-file=janusLogs.txt \
		--dev

run-janus-https:
	@ printf "\nRunning Janus...\n\n"

	go run `pwd`/main.go \
		--htmlcoin-rpc=http://${test_user}:${test_user_passwd}@0.0.0.0:4889 \
		--htmlcoin-network=auto \
		--bind=0.0.0.0 \
		--port=24889 \
		--accounts=`pwd`/docker/standalone/myaccounts.txt \
		--log-file=janusLogs.txt \
		--dev \
		--https-key https/key.pem \
		--https-cert https/cert.pem

test_user = htmlcoin
test_user_passwd = testpasswd

# Runs docker container of htmlcoin locally and starts htmlcoind inside of it
run-htmlcoin:
	@ printf "\nRunning htmlcoin...\n\n"
	@ printf "\n(1) Starting container...\n\n"
	docker run ${htmlcoin_container_flags} htmlcoin/htmlcoin htmlcoind ${htmlcoind_flags} > /dev/null

	@ printf "\n(2) Importing test accounts...\n\n"
	@ sleep 3
	docker cp ${shell pwd}/docker/fill_user_account.sh ${htmlcoin_container_name}:.

	@ printf "\n(3) Filling test accounts wallets...\n\n"
	docker exec ${htmlcoin_container_name} /bin/sh -c ./fill_user_account.sh > /dev/null
	@ printf "\n... Done\n\n"

seed-htmlcoin:
	@ printf "\n(2) Importing test accounts...\n\n"
	docker cp ${shell pwd}/docker/fill_user_account.sh ${htmlcoin_container_name}:.

	@ printf "\n(3) Filling test accounts wallets...\n\n"
	docker exec ${htmlcoin_container_name} /bin/sh -c ./fill_user_account.sh
	@ printf "\n... Done\n\n"

htmlcoin_container_name = test-chain

# TODO: Research -v
htmlcoin_container_flags = \
	--rm -d \
	--name ${htmlcoin_container_name} \
	-v ${shell pwd}/dapp \
	-p 4889:4889

# TODO: research flags
htmlcoind_flags = \
	-regtest \
	-rpcbind=0.0.0.0:4889 \
	-rpcallowip=0.0.0.0/0 \
	-logevents \
	-addrindex \
	-reindex \
	-txindex \
	-rpcuser=${test_user} \
	-rpcpassword=${test_user_passwd} \
	-deprecatedrpc=accounts \
	-printtoconsole

# Starts continuously printing htmlcoin container logs to the invoking terminal
follow-htmlcoin-logs:
	@ printf "\nFollowing htmlcoin logs...\n\n"
		docker logs -f ${htmlcoin_container_name}

open-htmlcoin-bash:
	@ printf "\nOpening htmlcoin bash...\n\n"
		docker exec -it ${htmlcoin_container_name} bash

# Stops docker container of htmlcoin
stop-htmlcoin:
	@ printf "\nStopping htmlcoin...\n\n"
		docker kill `docker container ps | grep ${htmlcoin_container_name} | cut -d ' ' -f1` > /dev/null
	@ printf "\n... Done\n\n"

restart-htmlcoin: stop-htmlcoin run-htmlcoin

submodules:
	git submodules init

# Run openzeppelin tests, Janus/HTMLCOIN needs to already be running
openzeppelin:
	cd testing && make openzeppelin

# Run openzeppelin tests in docker
# Janus and HTMLCOIN need to already be running
openzeppelin-docker:
	cd testing && make openzeppelin-docker

# Run openzeppelin tests in docker-compose
openzeppelin-docker-compose:
	cd testing && make openzeppelin-docker-compose
