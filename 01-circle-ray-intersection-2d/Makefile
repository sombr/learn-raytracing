all:
	GOOS=js GOARCH=wasm go build -o target/main.wasm
server:
	cd target && python3 -m http.server 9000
