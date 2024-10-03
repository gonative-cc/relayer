
setup:
	@cd .git/hooks; ln -s -f ../../scripts/git-hooks/* ./


out:
	mkdir out

.git/hooks/pre-commit: setup

build: out .git/hooks/pre-commit
	go build -o ./out ./cmd/rly-pera

run:
	@if [ ! -f .env ]; then echo "Create .env file by copying and updating .env.example"; exit 1; fi
	@./out/rly-pera start --block 8713586

clean:
	rm -rf out

# used as pre-commit
lint-git:
	@git diff --name-only --cached | grep  -E '\.go$$' | xargs revive
	@git diff --name-only --cached | grep  -E '\.md$$' | xargs markdownlint-cli2 ./NONE

# lint changed files
lint:
	@git diff --name-only | grep  -E '\.go$$' | xargs revive
	@git diff --name-only | grep  -E '\.md$$' | xargs markdownlint-cli2 ./NONE

lint-fix:
	@git diff --name-only  | grep  -E '\.md$$' | xargs markdownlint-cli2 --fix ./NONE

lint-all:
	@revive ./...

lint-gofmt-fix:
	@find -name "*.go" -exec gofmt -w -s {} \;

.PHONY: build run clean setup

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
