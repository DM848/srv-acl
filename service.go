package aclsrv

import (
	"errors"
	"strings"
	"sync"
)

// Service holds detail needed + round robin to load balance requests
// assumption: a Service object never exists if there are no addresses for it
type Service struct {
	sync.RWMutex
	Name      string   `json:"name"`
	Addresses []string `json:"addresses"`
	rri       int      // round robin index
}

func (s *Service) Empty() bool {
	return len(s.Addresses) == 0 || s.Name == ""
}

// GetAddress returns a <ip:port> string which is accessible within the cluster
func (s *Service) GetAddress() (adr string) {
	s.Lock()
	defer s.Unlock()

	adr = s.Addresses[s.rri]
	s.rri = (s.rri + 1) % len(s.Addresses)

	return
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
