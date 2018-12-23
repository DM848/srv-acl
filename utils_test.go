package aclsrv

import (
	"testing"
)

func check(t *testing.T, srv, path string) {
	got, err := getServiceName(path)
	if err != nil {
		t.Error(err)
	} else if srv != got {
		t.Errorf("extracted incorrect service name. Got %s, wants %s", got, srv)
	}
}

func TestGetServiceName(t *testing.T) {
	var srv string
	var path string

	srv = "test"
	path = "/" + srv
	check(t, srv, path)

	srv = "test-api"
	path = "/" + srv
	check(t, srv, path)

	srv = "test-api"
	path = "/" + srv + "/param1/param2"
	check(t, srv, path)

	srv = "test-api"
	path = "/" + srv + "?p=true"
	check(t, srv, path)

	srv = "test-api"
	path = "/" + srv + "?p=true/p2"
	check(t, srv, path)

	srv = ""
	path = "/" + srv
	if _, err := getServiceName(path); err == nil {
		t.Error(err)
	}

	srv = ""
	path = "/" + srv + "?p=true/p2"
	if _, err := getServiceName(path); err == nil {
		t.Error(err)
	}

}
