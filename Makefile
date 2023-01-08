.PHONY: updater
updater:
	export GH_USERNAME="anfelo" && \
	export GH_REPO="hello-tauri" && \
	export PRIVATE_BASE_URL="http://localhost:8080" && \
	reflex -s go run cmd/server/main.go
