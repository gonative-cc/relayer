
setup:
	@cd .git/hooks; ln -s ../../scripts/git-hooks/* ./


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

lint:
	@revive ./...

.PHONY: build run clean setup
