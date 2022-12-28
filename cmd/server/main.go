package main

import (
	"fmt"

	transportHttp "github.com/anfelo/go-updater/internal/transport/http"
)

// Run - responsible for the instantiation
// and startup of our go application
func Run() error {
	fmt.Println("Starting up the application")

	httpHandler := transportHttp.NewHandler()
	if err := httpHandler.Serve(); err != nil {
		return err
	}

	return nil
}

func main() {
	fmt.Println("Go Electron Updater")
	if err := Run(); err != nil {
		fmt.Println(err)
	}
}
