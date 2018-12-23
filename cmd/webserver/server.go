package main

import (
	"log"
	"net/http"
	"os"

	"aclsrv"
	"github.com/julienschmidt/httprouter"
)

func main() {
	port := os.Getenv("WEB_SERVER_PORT")
	if port == "" {
		panic("missing environment variable WEB_SERVER_PORT")
	}



	// ACL state to hold all configs and such
	ACLState := aclsrv.NewState()

	router := httprouter.New()

	consul, err := aclsrv.NewConsul(nil, "./service.json")
	if err != nil {
		panic(err)
	}
	consul.HealthCheck(router)
	consul.Register()

	aclsrv.SetupRoutes(router, ACLState)



	log.Fatal(http.ListenAndServe(":" + port, router))
}
