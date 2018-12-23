package aclsrv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/julienschmidt/httprouter"
)
const (
	ConsulKVPrefix = "srv-acl_"
)

func NewState() *State {
	return &State{
		httpClient: http.DefaultClient,
	}
}

type State struct {
	sync.RWMutex

	// updated through service discovery
	// a service might exist here, while not in th ACL.
	// That simply means the service has no ACL configuration and that everyone has access
	Services []*Service `json:"services"`

	// ACL
	ACL []*ACLEntry `json:"ACLEntries"`

	UserScripts []*Service `json:"user_scripts"`

	httpClient *http.Client
}

// Get service if it exists
func (s *State) Service(name string) (srv *Service) {
	s.RLock()
	defer s.RUnlock()

	for _, srv = range s.Services {
		if !srv.Empty() && name == srv.Name {
			return
		}
	}

	return nil
}

// Get user script if it exists
func (s *State) UserScript(name string) (srv *Service) {
	s.RLock()
	defer s.RUnlock()

	for _, srv = range s.UserScripts {
		if !srv.Empty() && name == srv.Name {
			return
		}
	}

	return nil
}

func (s *State) ServiceACL(srv *Service) (entry *ACLEntry) {
	s.RLock()
	defer s.RUnlock()

	for _, entry = range s.ACL {
		if !entry.Empty() && srv.Name == entry.Service {
			return
		}
	}

	return nil
}

func (s *State) APIHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	response := &JSend{}
	defer func(response *JSend){
		response.write(w)
	}(response)

	path := ps.ByName(APIPathID)
	srvName, err := getServiceName(path)
	if err != nil {
		response.Status = JSendFail
		response.Message = "unable to get service name from your request. Error: " + err.Error()
		return
	}

	// check if such a service exists
	srv := s.Service(srvName)

	// verify that this is a public route/service
	if srv == nil {
		response.Status = JSendFail
		response.Message = "service was not found or does not exist as an endpoint yet"
		response.HTTPCode = 404
		return
	}

	// verify JWT signature and get user info
	user, length, err := getJWTUser(r)
	ignoreCheck := length > 0
	if err != nil && ignoreCheck {
		response.Status = JSendFail
		response.Message = "issue with JWT. Error: " + err.Error()
		return
	}

	// verify permissions / ACL
	if acl := s.ServiceACL(srv); acl != nil && ignoreCheck && !acl.HasAccess(user) {
		response.Status = JSendFail
		response.Message = "You do not have access to this service. Error: " + err.Error()
		return
	}

	// recreate request
	previous := "/" + srvName
	addr := "http://" + srv.GetAddress() + path[len(previous):]
	internalReq, err := http.NewRequest(r.Method, addr, r.Body)
	if err != nil {
		response.Status = JSendError
		response.Message = err.Error()
		return
	}

	internalReq.Header.Set("Accept", "application/json")
	resp, err := s.httpClient.Do(internalReq)
	if err != nil {
		response.Status = JSendError
		response.Message = err.Error()
		return
	}

	// handle response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		response.Status = JSendError
		response.Message = err.Error()
		return
	}

	// success
	response.Status = JSendSuccess
	response.Data = body
	response.HTTPCode = resp.StatusCode
}

func (s *State) ScriptHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	path := ps.ByName(APIPathID)
	srvName, err := getServiceName(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check if such a service exists
	var srv *Service
	if srv = s.UserScript(srvName); srv == nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// verify we have created an acceptable URL
	addr := "http://" + srv.GetAddress() + path[len("/" + srvName):]
	if _, err = url.Parse(addr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// we need to buffer the body if we want to read it here and send it
	// in the request.
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// you can reassign the body if you need to parse it as multipart
	r.Body = ioutil.NopCloser(bytes.NewReader(body))
	proxyReq, err := http.NewRequest(r.Method, addr, bytes.NewReader(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	proxyReq.Header = r.Header

	resp, err := s.httpClient.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	for k, v := range resp.Header {
		w.Header().Set(k, v[0])
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.Write(body)
}


func (s *State) WatchAliveServicesHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	response := &JSend{}
	defer func(response *JSend){
		response.write(w)
	}(response)

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(body, s)
	if err != nil {
		log.Fatal(err)
	}

	data, _ := json.Marshal(s)
	fmt.Println(string(data))
}