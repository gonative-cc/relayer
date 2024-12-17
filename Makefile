
setup:
	@cd .git/hooks; ln -s -f ../../scripts/git-hooks/* ./


out:
	mkdir out

.git/hooks/pre-commit: setup

build: out .git/hooks/pre-commit
	go build -o ./out ./cmd/*

start:
	@if [ ! -f .env ]; then echo "Create .env file by copying and updating .env.example"; exit 1; fi
	@./out/native-ika start --block 8713586

clean:
	rm -rf out

# used as pre-commit
lint-git:
	@files=$$(git diff --name-only --cached | grep  -E '\.go$$' | xargs -r gofmt -l); if [ -n "$$files" ]; then echo $$files;  exit 101; fi
	@git diff --name-only --cached | grep  -E '\.go$$' | xargs -r revive
	@git diff --name-only --cached | grep  -E '\.md$$' | xargs -r markdownlint-cli2

# lint changed files
lint:
	@files=$$(git diff --name-only | grep  -E '\.go$$' | xargs -r gofmt -l); if [ -n "$$files" ]; then echo $$files;  exit 101; fi
	@git diff --name-only | grep  -E '\.go$$' | xargs -r revive
	@git diff --name-only | grep  -E '\.md$$' | xargs -r markdownlint-cli2

lint-all: lint-fix-go-all
	@revive ./...

lint-fix-all: lint-fix-go-all

lint-fix-go-all:
	@gofmt -w -s -l .


.PHONY: build start clean setup
.PHONY: lint lint-all lint-fix-all lint-fix-go-all

###############################################################################
##                                   Tests                                   ##
###############################################################################

TEST_COVERAGE_PROFILE=coverage.txt
TEST_TARGETS := test-unit test-unit-cover test-race
test-unit: ARGS=-timeout=10m -tags='$(UNIT_TEST_TAGS)'
test-unit-cover: ARGS=-timeout=10m -tags='$(UNIT_TEST_TAGS)' -coverprofile=$(TEST_COVERAGE_PROFILE) -covermode=atomic
test-race: ARGS=-timeout=10m -race -tags='$(TEST_RACE_TAGS)'
$(TEST_TARGETS): run-tests

run-tests:
ifneq (,$(shell which tparse 2>/dev/null))
	@go test -mod=readonly -json $(ARGS) ./... | tparse
else
	@go test -mod=readonly $(ARGS) ./...
endif

cover-html: test-unit-cover
	@echo "--> Opening in the browser"
	@go tool cover -html=$(TEST_COVERAGE_PROFILE)

###############################################################################
##                                Infrastructure                             ##
###############################################################################

BITCOIND_SNAPSHOT=$(shell pwd)/contrib/snapshot
BITCOIND_DATA=$(shell pwd)/contrib/regtest

run-bitcoind:
	@docker run -v $(BITCOIND_DATA):/bitcoin/.bitcoin --name=bitcoind-node -d \
        -p 18444:8333 \
        -p 127.0.0.1:18443:8332 \
        -e REGTEST=1 \
        -e DISABLEWALLET=0 \
        -e PRINTTOCONSOLE=1 \
        -e RPCUSER=mysecretrpcuser \
        -e RPCPASSWORD=mysecretrpcpassword \
        kylemanna/bitcoind
	@sleep 1
	@docker exec -it bitcoind-node bitcoin-cli -regtest -rpcport=8332 loadwallet "nativewallet"

create-bitcoind: snapshot run-bitcoind

start-bitcoind:
	@docker start bitcoind-node

stop-bitcoind:
	@docker stop bitcoind-node

delete-bitcoind:
	@docker rm -f bitcoind-node

restart-bitcoind: delete-bitcoind create-bitcoind

snapshot:
	@rm -rf $(BITCOIND_DATA)
	@cp -rf $(BITCOIND_SNAPSHOT) $(BITCOIND_DATA)




