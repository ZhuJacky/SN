// Package main provides ...
package main

import (
	"log"
	"net/http"
	"strings"
)

func main() {
	// http server
	log.Println("echo ip tools running :8089")

	http.HandleFunc("/latest/meta-data/local-ipv4", HandleEchoIP)
	http.HandleFunc("/latest/meta-data/instance-type", HandleInstanceType)
	log.Fatal(http.ListenAndServe(":8089", nil))
}

// HandleInstanceType doc
func HandleInstanceType(w http.ResponseWriter, r *http.Request) {
	// NOTE: for development
	w.Write([]byte("c4.large"))
}

func HandleEchoIP(w http.ResponseWriter, r *http.Request) {
	index := strings.LastIndex(r.RemoteAddr, ":")
	w.Write([]byte(r.RemoteAddr[:index]))
}
