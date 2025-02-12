package server

import (
	"fmt"
	"github.com/arl/statsviz"
	example "github.com/arl/statsviz/_example"
	"log"
	"net/http"
)

func NewHttpDebugging() {
	// Force the GC to work to make the plots "move".
	go example.Work()

	// Register a Statsviz server on the default mux.
	_ = statsviz.Register(http.DefaultServeMux)

	fmt.Println("Http debugging started at http://localhost:8080/debug/statsviz/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
