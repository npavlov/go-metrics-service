package main

import (
	"fmt"
	"github.com/npavlov/go-metrics-service/internal/flags"
	"github.com/npavlov/go-metrics-service/internal/server/handlers/render"
	"github.com/npavlov/go-metrics-service/internal/server/handlers/retrieve"
	"github.com/npavlov/go-metrics-service/internal/server/handlers/update"
	"github.com/npavlov/go-metrics-service/internal/server/router"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"github.com/npavlov/go-metrics-service/internal/types"
	"log"
	"net/http"
)

func main() {
	parseFlags()
	flags.VerifyFlags()

	var memStorage storage.Repository = storage.NewMemStorage()

	handlers := types.Handlers{
		UpdateHandler:   update.GetUpdateHandler(memStorage),
		RetrieveHandler: retrieve.GetRetrieveHandler(memStorage),
		RenderHandler:   render.GetRenderHandler(memStorage),
	}

	r := router.GetRouter(handlers)

	// Launching server at :8080
	fmt.Printf("Server started at %s", flagRunAddr)
	log.Fatal(http.ListenAndServe(flagRunAddr, r))
}
