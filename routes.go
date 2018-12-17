package aclsrv

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
)

const (
	APIPathID = "path"
)

type UserID uint64
type Permission uint32

// Service holds detail needed + round robin to load balance requests
// assumption: a Service object never exists if there are no pods for it
type Service struct {
	sync.RWMutex
	Name string
	pods []string
	rri int // round robin index
}

func (s *Service) Empty() bool {
	return len(s.pods) == 0
}

// GetAddress returns a <ip:port> string which is accessible within the cluster
func (s *Service) GetAddress() (adr string) {
	s.Lock()
	defer s.Unlock()

	adr = s.pods[s.rri]
	s.rri = (s.rri + 1) & len(s.pods)

	return
}

type User struct {
	UserID UserID
	Permissions []Permission
}

type ACL struct {
	Service *Service
	AllowedPermissions []Permission
	AllowedUserIDs []UserID
	BlockedUserIDs []UserID
	LastUpdated int64 // unix
}

func (a *ACL) HasAccess(user *User) bool {
	// check if blocked
	for _, permission := user.Permissions {
		for _, accepted := a.AllowedPermissions {
			if
		}
	}
}

// getServiceName extract the service name from the uri path, after the /api prefix
func getServiceName(apiSuffix string) (service string, err error) {
	if apiSuffix[0] == '/' {
		apiSuffix = apiSuffix[1:]
	}
	if len(apiSuffix) == 0 {
		err = errors.New("missing service name")
		return
	}

	slash := strings.IndexByte(apiSuffix, '/')
	questionmark := strings.IndexByte(apiSuffix, '?')

	service = apiSuffix // fallback
	if slash >= 0 {
		service = apiSuffix[:slash]
	}
	if questionmark >= 0 {
		service = apiSuffix[:questionmark]
	}

	if len(service) == 0 {
		err = errors.New("missing service name")
	}
	return
}

func isService(srv string) bool {
	// TODO: implement a list over all the available services!
	return true
}

func apiHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	srv, err := getServiceName(ps.ByName(APIPathID))
	if err != nil {
		msg := &JSend{
			Status: JSendFail,
			Message: "unable to get service name from your request. Error: " + err.Error(),
		}
		msg.write(w)
		return
	}

	// verify that this is a route in ACL and check if it needs JWT
	if isService()


	fmt.Fprintf(w, srv)
}

func SetupRoutes(router *httprouter.Router) {
	router.GET("/health", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Fprintf(w, `{"status":"ok"}`)
	})
	router.GET("/dashboard", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Fprintf(w, `TODO: where is dashboard static files stored?`)
	})


	accepts := []string{
		http.MethodGet,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodOptions,
		http.MethodPost,
		http.MethodPut,
	}
	for _, method := range accepts {
		router.Handle(method, "/api/*" + APIPathID, apiHandler)
	}
}
