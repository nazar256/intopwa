.PHONY: dev
dev:
	wrangler dev

.PHONY: build
build:
	go run github.com/syumai/workers/cmd/workers-assets-gen@v0.27.0
	rm -f build/app.wasm
	tinygo build -tags cloudflare -o ./build/app.wasm -target wasm -gc=leaking -no-debug ./main.go
	npx wasm-opt -Os ./build/app.wasm -o ./build/app.wasm

.PHONY: deploy
deploy:
	npx wrangler deploy

gen:
	go generate ./...