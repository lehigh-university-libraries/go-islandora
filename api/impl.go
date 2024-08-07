package api

import (
	"encoding/json"
	"net/http"

	islandoraModel "github.com/lehigh-university-libraries/go-islandora/model"
)

// ensure that we've conformed to the `ServerInterface` with a compile-time check
var _ ServerInterface = (*Server)(nil)

type Server struct{}

func NewServer() Server {
	return Server{}
}

// (POST /upload)
func (Server) PostUpload(w http.ResponseWriter, r *http.Request) {

	// TODO: transform the vanilla CSV into workbench CSV
	resp := IslandoraObject{
		Title: &islandoraModel.GenericField{
			islandoraModel.Generic{
				Value: "foo",
			},
		},
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
