out:
	mkdir out

build: out
	go build -o ./out ./cmd/rly-pera

run:
	@if [ ! -f .env ]; then echo "Create .env file by copying and updating .env.example"; exit 1; fi
	@./out/rly-pera start --block 8713586

clean:
	rm -rf out

.PHONY: build run clean
