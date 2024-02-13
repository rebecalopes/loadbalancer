package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	servers := []Server{
		newSimpleServer("https://pt-br.facebook.com/"),
		newSimpleServer("https://twitter.com/"),
		newSimpleServer("https://www.instagram.com/"),
	}

	lb := newLoadBalancer("8000", servers)
	handleRedirect := func(rw http.ResponseWriter, req *http.Request) {
		lb.serverProxy(rw, req)
	}
	http.HandleFunc("/", handleRedirect)

	fmt.Printf("serving requests al localhost: %s\n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}

type Server interface {
	Address() string
	isAlive() bool
	Server(rw http.ResponseWriter, r *http.Request)
}

type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

func newSimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)
	handleErr(err)

	return &simpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

type loadbalancer struct {
	port            string
	roundRobinCount int
	servers         []Server
}

func newLoadBalancer(port string, servers []Server) *loadbalancer {
	return &loadbalancer{
		port:            port,
		roundRobinCount: 0,
		servers:         servers,
	}
}

func handleErr(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

func (s *simpleServer) Address() string {
	return s.addr
}

func (s *simpleServer) isAlive() bool {
	return true
}

func (s *simpleServer) Server(rw http.ResponseWriter, req *http.Request){
	s.proxy.ServeHTTP(rw, req)
}

func (lb *loadbalancer) getNextAvalibleServer() Server {
	server := lb.servers[lb.roundRobinCount%len(lb.servers)]

	for !server.isAlive(){
		lb.roundRobinCount++
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]
	}

	lb.roundRobinCount++
	return server
}

func (lb *loadbalancer) serverProxy(rw http.ResponseWriter, req *http.Request) {
	targetServer := lb.getNextAvalibleServer()
	fmt.Printf("forwarding request to address %q\n", targetServer.Address())
	targetServer.Server(rw, req)
}
