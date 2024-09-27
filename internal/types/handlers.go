package types

import "net/http"

type Handlers struct {
	UpdateHandler   func(http.ResponseWriter, *http.Request)
	RetrieveHandler func(http.ResponseWriter, *http.Request)
	RenderHandler   func(http.ResponseWriter, *http.Request)
}
