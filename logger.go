package aclsrv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	log2 "log"
	"net/http"
	"time"
)

// Java log levels as an integer
const (
	LogLvlWarn   = 900
	LogLvlINFO   = 800
	LogLvlFINEST = 300
)

type ACLLogEntry struct {
	Name  string `json:"service"`
	Level int    `json:"level"`
	Info  string `json:"info"`
}

type LEapi struct {
	OriginalURL string      `json:"original_url"`
	ProxiedURL  string      `json:"proxied_url"`
	IP          string      `json:"ip"`
	ReqHeader   http.Header `json:"req_header,omitempty"`
	ResHeader   http.Header `json:"res_header,omitempty"`
}

func (e *LEapi) String() string {
	data, err := json.Marshal(e)
	if err != nil {
		return ""
	}

	return string(data)
}

var _ fmt.Stringer = (*LEapi)(nil)

var logClient = http.DefaultClient

// contacts the logging service
func logger(level int, info fmt.Stringer) (dbIndex int) {
	logClient.Timeout = time.Duration(2 * time.Second)

	entry := &ACLLogEntry{
		Name:  "acl",
		Level: level,
		Info:  info.String(),
	}
	if entry.Info == "" {
		log2.Print("empty entry")
		return -1
	}
	data, err := json.Marshal(entry)
	if err != nil {
		log2.Fatal(err)
		return -1
	}

	var body io.ReadCloser
	body = ioutil.NopCloser(bytes.NewBuffer(data))
	req, err := http.NewRequest(http.MethodPost, "http://logger:8888/set", body)
	if err != nil {
		log2.Fatal(err)
		return -1
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")
	resp, err := logClient.Do(req)
	if err != nil {
		log2.Fatal(err)
		return -1
	}

	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log2.Fatal(err)
		return -1
	}

	fmt.Println("logger response: " + string(data))

	return -1 // TODO: use the response
}
