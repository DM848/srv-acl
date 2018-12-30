package aclsrv

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

const (
	APIPathID = "path"
)

type UserLevel struct {
	Role       string     `json:"role"`
	Permission Permission `json:"permission"`
}

type ACLInfo struct {
	UserLevels []*UserLevel `json:"user_levels,omitempty"`
	ACLConfig  []*ACLEntry  `json:"acl_endpoints,omitempty"`
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, PATCH, OPTIONS, PUT, DELETE")

	// TODO: what are the required jolie headers?
	//(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func SetupRoutes(router *httprouter.Router, ACLState *State) {
	router.GET("/configuration", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		setupResponse(&w, r)

		response := &JSend{
			HTTPCode: http.StatusOK,
		}
		defer func(response *JSend) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(response.HTTPCode)
			response.write(w)
		}(response)

		list := &ACLInfo{
			UserLevels: ACLState.PermissionDefaults,
			ACLConfig:  ACLState.ACL,
		}

		// add services without ACL entry
		ACLState.RLock()
		for i := range ACLState.Services {
			exists := false
			for j := range list.ACLConfig {
				exists = ACLState.Services[i].Name == list.ACLConfig[j].Service
				if exists {
					break
				}
			}

			if !exists {
				list.ACLConfig = append(list.ACLConfig, &ACLEntry{
					Service: ACLState.Services[i].Name,
				})
			}
		}
		ACLState.RUnlock()

		data, err := json.Marshal(list)
		if err != nil {
			response.Status = JSendError
			response.Message = "unable to unmarshal list. Error: " + err.Error()
			response.HTTPCode = http.StatusInternalServerError
			return
		}

		response.Status = JSendSuccess
		response.Data = data
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
		router.Handle(method, "/api/*"+APIPathID, ACLState.APIHandler)
	}
	for _, method := range accepts {
		router.Handle(method, "/script/*"+APIPathID, ACLState.ScriptHandler)
	}
}
