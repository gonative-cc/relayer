
setup-hooks:
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

gen-mocks:
	go run github.com/vektra/mockery/v2

###############################################################################
##                                Infrastructure                             ##
###############################################################################

WALLET="nativewallet"

bitcoind-init:
	@rm -rf ./contrib/bitcoind-data
	@cp -rf ./contrib/bitcoind-snapshot contrib/bitcoind-data

bitcoind-load-wallet:
	@docker exec -it bitcoind-node bitcoin-cli -regtest loadwallet $(WALLET)


# if you want to start all containers that we defined in the docker-compose, then just run
# docker compose up
docker-sui-start:
	@cd ./contrib; docker compose up sui-node

# make sure you start docker container is started
docker-sui-connect:
	@cd ./contrib; docker compose exec sui-node bash
# or: docker exec -it sui-node bash

docker-bitcoind-connect:
	@cd ./contrib; docker compose exec bitcoind bash
