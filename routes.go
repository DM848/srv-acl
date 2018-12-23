package aclsrv

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

const (
	APIPathID = "path"
)

func SetupRoutes(router *httprouter.Router, ACLState*State) {
	router.GET("/dashboard", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Fprintf(w, `TODO: where is dashboard static files stored?`)
	})

	router.POST("/consul/services/change", ACLState.WatchAliveServicesHandler)

	// setup
	accepts := []string{
		http.MethodGet,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodOptions,
		http.MethodPost,
		http.MethodPut,
	}
	for _, method := range accepts {
		router.Handle(method, "/api/*" + APIPathID, ACLState.APIHandler)
	}
	for _, method := range accepts {
		router.Handle(method, "/script/*" + APIPathID, ACLState.ScriptHandler)
	}
}
