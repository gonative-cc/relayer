
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

.PHONY: build run clean setup
