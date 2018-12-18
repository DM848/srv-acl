package aclsrv

import (
	"bytes"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const ConsulAddress = "http://consul-node:8500"

type ConsulCheck struct {
	HTTP string `json:"HTTP,omitempty"`
	Interval string `json:"interval,omitempty"`
	Method string `json:"method,omitempty"`
	Timeout string `json:"timeout,omitempty"`
	Args []string `json:"args,omitempty"`
}
type ConsulWeights struct {
	Passing int `json:"passing"`
	Warning int `json:"warning"`
}

type ConsulSrvDef struct {
	// Name - Required - Specifies the logical name of the service. Many service instances may share the same logical service name.
	Name string `json:"name"`

	// ID Specifies a unique ID for this service. This must be unique per agent. This defaults to the Name parameter if not provided.
	ID string `json:"ID,omitempty"`

	// Tags Specifies a list of tags to assign to the service. These tags can be used for later filtering and are exposed via the APIs.
	Tags []string `json:"tags,omitempty"`

	// Address Specifies the address of the service. If not provided, the agent's address is used as the address for the service during DNS queries.
	Address string `json:"address,omitempty"`

	// Meta Specifies arbitrary KV metadata linked to the service instance.
	Meta map[string]string `json:"meta,omitempty"`

	// Port Specifies the port of the service.
	Port uint16 `json:"port,omitempty"`

	// Kind The kind of service. Defaults to "" which is a typical consul service. This value may also be "connect-proxy" for services that are Connect-capable proxies representing another service.
	Kind string `json:"kind,omitempty"`

	// Proxy (Proxy: nil) - From 1.2.3 on, specifies the configuration for a Connect proxy instance. This is only valid if Kind == "connect-proxy". See the Proxy documentation for full details.
	Proxy interface{} `json:"proxy,omitempty"`

	// Connect (Connect: nil) - Specifies the configuration for Connect. See the Connect Structure section below for supported fields.
	Connect interface{} `json:"connect,omitempty"`

	// Check (Check: nil) - Specifies a check. Please see the check documentation for more information about the accepted fields. If you don't provide a name or id for the check then they will be generated. To provide a custom id and/or name set the CheckID and/or Name field.
	Check *ConsulCheck `json:"check,omitempty"`

	// Checks (array<Check>: nil) - Specifies a list of checks. Please see the check documentation for more information about the accepted fields. If you don't provide a name or id for the check then they will be generated. To provide a custom id and/or name set the CheckID and/or Name field. The automatically generated Name and CheckID depend on the position of the check within the array, so even though the behavior is deterministic, it is recommended for all checks to either let consul set the CheckID by leaving the field empty/omitting it or to provide a unique value.
	Checks []*ConsulCheck `json:"checks,omitempty"`

	// EnableTagOverride (bool: false) - Specifies to disable the anti-entropy feature for this service's tags. If EnableTagOverride is set to true then external agents can update this service in the catalog and modify the tags. Subsequent local sync operations by this agent will ignore the updated tags. For instance, if an external agent modified both the tags and the port for this service and EnableTagOverride was set to true then after the next sync cycle the service's port would revert to the original value but the tags would maintain the updated value. As a counter example, if an external agent modified both the tags and port for this service and EnableTagOverride was set to false then after the next sync cycle the service's port and the tags would revert to the original value and all modifications would be lost.
	EnableTagOverride bool `json:"enable_tag_override,omitempty"`

	// Weights (Weights: nil) - Specifies weights for the service. Please see the service documentation for more information about weights. If this field is not provided weights will default to {"Passing": 1, "Warning": 1}.
	Weights *ConsulWeights `json:"weights,omitempty"`
}

func NewConsul(client *http.Client, file string) (*consul, error) {
	if client == nil {
		client = http.DefaultClient
	}

	c := &consul{
		client: client,
	}

	return c, c.LoadSrvDef(file)
}

type consul struct {
	srv ConsulSrvDef
	registerred bool
	client *http.Client
}

func (c *consul) LoadSrvDef(file string) (err error) {
	var data []byte
	data, err = ioutil.ReadFile(file)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &c.srv)
	if err != nil {
		return
	}

	// set ID if empty
	if c.srv.ID == "" {
		podID := os.Getenv("HOSTNAME")
		if podID == "" {
			podID = c.srv.Name
		}

		c.srv.ID = podID
	}

	// set address
	c.srv.Address = os.Getenv("MY_POD_IP")

	return
}

func (c *consul) HealthCheck(router *httprouter.Router) {
	router.GET("/health", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		response := &JSend{}
		response.Status = JSendSuccess
		response.Data = []byte(`{"status":"ok"}`)

		response.write(w)
	})
}

func (c *consul) Register() {
	if c.srv.Name == "" {
		panic("have not loaded service definition from file")
	}

	data, err := json.Marshal(&c.srv)
	if err != nil {
		panic(err)
	}

	const url = "http://consul-node:8500/v1/agent/service/register"
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	c.registerred = resp.StatusCode == http.StatusOK
}
func (c *consul) Deregister() {
	if c.srv.Name == "" {
		panic("have not loaded service definition from file")
	}

	const url = "http://consul-node:8500/v1/agent/service/deregister/"
	req, err := http.NewRequest(http.MethodPut, url + c.srv.ID, nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	c.registerred = !(resp.StatusCode == http.StatusOK)
}