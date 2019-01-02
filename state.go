package aclsrv

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
	"github.com/lestrrat-go/jwx/jwk"
)

const (
	ConsulKVPrefix = "srv-acl_"
)

func NewState() *State {
	return &State{
		httpClient: http.DefaultClient,
	}
}

type ACLConfigEntry struct {
	Key string      `json:"key"`
	Val interface{} `json:"val"`
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

	PermissionDefaults []*UserLevel `json:"ACLRolesPermission"`

	Config []ACLConfigEntry `json:"config"`

	httpClient *http.Client

	jwksMu sync.RWMutex
	jwks   *jwk.Set
}

func (s *State) lookupConfig(key string) string {
	s.RLock()
	defer s.RUnlock()

	for i := range s.Config {
		if s.Config[i].Key == key {
			return fmt.Sprint(s.Config[i].Val)
		}
	}

	return ""
}

func (s *State) getJWK(kid string) (interface{}, error) {
	s.jwksMu.RLock()
	if key := s.jwks.LookupKeyID(kid); len(key) == 1 {
		s.jwksMu.RUnlock()
		return key[0].Materialize()
	}
	s.jwksMu.RUnlock()

	// get fresh keys
	set, err := jwk.FetchHTTP(jwksURL)
	if err != nil {
		return nil, err
	}

	// add keys to cache
	s.jwksMu.Lock()
	defer s.jwksMu.Unlock()
	for i := range set.Keys {
		k := set.Keys[i].KeyID()
		if key := s.jwks.LookupKeyID(k); len(key) == 1 {
			continue
		}

		s.jwks.Keys = append(s.jwks.Keys, set.Keys[i])
	}

	if key := s.jwks.LookupKeyID(kid); len(key) == 1 {
		return key[0].Materialize()
	}

	return nil, errors.New("unable to find key")
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
	setupResponse(&w, r)

	response := &JSend{
		HTTPCode: http.StatusOK,
	}
	var addr string // proxied addr
	var user *User
	defer func(response *JSend) {
		response.write(w)

		go logger(LogLvlINFO, &LEapi{
			IP:          r.Host,
			User:        user,
			OriginalURL: r.URL.String(),
			ProxiedURL:  addr,
			Err: response.Message,
		})
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
	tokenStr := getJWT(r.Header)
	if tokenStr == "" {
		if s.lookupConfig("jwt") == "true" && srvName != "jolie-deployer" {
			response.Status = JSendFail
			response.Message = "Missing JWT in header. Supported fields: 'Authorization: Bearer <JWT>', 'jwt: <jwt>', 'JWT: <jwt>'"
			return
		} else {
			tokenStr = "a.b.c"
		}
	}
	user = &User{}
	// so.. right now we haven't found a proper way to deal with jolie-deployer in regards to
	// security. So I'm making the JWT token optional...
	//
	// This allows the ACL to check the actual permission of the jolie-deployer. Such that if those permissions are
	// ever added. You must be authenticated. Right now, the jolie deployer is hardcoded into the if else
	// to make it an exception. With this, at least we don't have to make every other service public as well.
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("unable to convert kid to string")
		}

		return s.getJWK(kid)
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if usrname := claims["cognito:username"]; usrname != nil {
			user.ID = usrname.(UserID)
		}
		if p := claims["cognito:groups"]; p != nil {
			arr := p.([]string)
			for i := range arr {
				if arr[i][:2] == "p:" {
					lvl, err := strconv.ParseUint(arr[i][2:], 10, 64)
					if err != nil {
						err = errors.New(err.Error() + " :::: Unable to extract permission level")
					} else {
						user.Permission = Permission(lvl)
					}
					break
				}
			}
		}
	}

	if (!token.Valid || err != nil) && s.lookupConfig("jwt") == "true" && srvName != "jolie-deployer" {
		response.Status = JSendFail
		response.Message = "issue with JWT. " + err.Error()

		if user.ID == "" {
			response.Message += " ::: also missing username"
		}

		return
	}

	// verify permissions / ACL
	// default: whitelist everyone if no ACL config is set for service
	if acl := s.ServiceACL(srv); acl != nil && !acl.HasAccess(user) {
		response.Status = JSendFail
		response.Message = "You do not have access to this service. Error: " + err.Error()
		return
	}

	// variable enforcement - see README.md
	urlValues := r.URL.Query()
	if s.lookupConfig("enforce") == "true" {
		var l int64
		r.Body, l, err = enforceJSONBodyParams(r.Body, user)
		if err != nil {
			response.Status = JSendFail
			response.Message = "Unable to handle the ACL enforced variables. Error: " + err.Error()
			return
		}
		r.ContentLength = l
		r.Header.Set("Content-Length", strconv.FormatInt(l, 10))
		enforceURLQueryParams(&urlValues, user) // TODO: review pointer
	}

	// recreate request
	previous := "/" + srvName
	addr = "http://" + srv.GetAddress() + path[len(previous):]
	urlQuery := urlValues.Encode()
	if urlQuery != "" {
		addr += "?" + urlQuery
	}
	internalReq, err := http.NewRequest(r.Method, addr, r.Body)
	if err != nil {
		response.Status = JSendError
		response.Message = err.Error()
		return
	}

	internalReq.Header = r.Header
	internalReq.Header.Set("Accept", "application/json")
	internalReq.Header.Del("Accept-Encoding")
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
	for k, v := range resp.Header {
		w.Header().Set(k, v[0])
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.Status = JSendSuccess
	response.Data = body
	response.InternalHTTPCode = resp.StatusCode

	go logger(LogLvlINFO, &LEapi{
		IP:          r.Host,
		OriginalURL: r.URL.String(),
		ProxiedURL:  addr,
	})
}

func (s *State) ScriptHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	setupResponse(&w, r)

	path := ps.ByName(APIPathID)
	srvName, err := getServiceName(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check if such a service exists
	var srv *Service
	if srv = s.UserScript(srvName); srv == nil {
		http.Error(w, "Unable to find user script for given token: "+srvName, http.StatusNotFound)
		return
	}

	// verify we have created an acceptable URL
	addr := "http://" + srv.GetAddress() + path[len("/"+srvName):]
	urlQuery := r.URL.Query().Encode()
	if urlQuery != "" {
		addr += "?" + urlQuery
	}
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
	defer func(response *JSend) {
		response.write(w)
	}(response)

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
		return
	}

	s.Lock()
	defer s.Unlock()
	err = json.Unmarshal(body, s)
	if err != nil {
		log.Fatal(err)
		return
	}

	// data, _ := json.Marshal(s)
	// fmt.Println(string(data))
}
