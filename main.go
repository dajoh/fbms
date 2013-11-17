package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

var config = LoadConfig("config.json")
var serverList = NewServerList(config)

func main() {
	http.HandleFunc("/list", listHandler)
	http.HandleFunc("/publish", publishHandler)
	log.Fatal(http.ListenAndServe(config.ListenAddress, nil))
}

func listHandler(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		badRequest(rw)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.Write(serverList.List())
}

func publishHandler(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		badRequest(rw)
		return
	}

	host := ""
	port := req.Header.Get("Fb-Port")

	if config.BeingProxied {
		host = req.Header.Get(config.RealIPHeader)
	} else {
		host, _, _ = net.SplitHostPort(req.RemoteAddr)
	}

	if port == "" {
		badRequest(rw)
		return
	}

	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		internalError(rw, err)
		return
	} else if len(payload) > config.MaxPayloadSize {
		badRequest(rw)
		return
	}

	err = serverList.Publish(net.JoinHostPort(host, port), payload)
	if err != nil {
		internalError(rw, err)
	}

	rw.Header().Set("Fb-Expire", fmt.Sprint(config.ExpireTime))
}

func badRequest(rw http.ResponseWriter) {
	http.Error(rw, "400 bad request", http.StatusBadRequest)
}

func internalError(rw http.ResponseWriter, err error) {
	http.Error(rw, err.Error(), http.StatusInternalServerError)
}
