package main

import (
	"fmt"

	"github.com/anfelo/go-updater/internal/datasource"
	transportHttp "github.com/anfelo/go-updater/internal/transport/http"
	"github.com/anfelo/go-updater/internal/updater"
)

// Run - responsible for the instantiation
// and startup of our go application
func Run() error {
	fmt.Println("Starting up the application")

	datasource := datasource.NewDatasource()
	updaterService := updater.NewService(datasource)

	httpHandler := transportHttp.NewHandler(updaterService)
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
