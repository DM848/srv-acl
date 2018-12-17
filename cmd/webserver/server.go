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
	router := httprouter.New()
	aclsrv.SetupRoutes(router)

	log.Fatal(http.ListenAndServe(":" + port, router))
}
