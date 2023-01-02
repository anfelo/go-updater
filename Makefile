.PHONY: updater
updater:
	export GH_USERNAME="anfelo" && \
	export GH_REPO="hello-tauri" && \
	reflex -s go run cmd/server/main.go
