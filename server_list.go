package main

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

type Server struct {
	timer   *time.Timer
	payload string
}

type ServerList struct {
	lock        sync.RWMutex
	config      *Config
	servers     map[string]*Server
	serverCount map[string]int
}

func NewServerList(config *Config) *ServerList {
	sl := new(ServerList)
	sl.config = config
	sl.servers = make(map[string]*Server)
	sl.serverCount = make(map[string]int)
	return sl
}

func (sl *ServerList) List() []byte {
	list := make(map[string]string)

	sl.lock.RLock()
	{
		for k, v := range sl.servers {
			list[k] = v.payload
		}
	}
	sl.lock.RUnlock()

	data, err := json.Marshal(list)
	if err != nil {
		log.Fatal(err)
	}

	return data
}

func (sl *ServerList) Publish(hostport string, payload []byte) error {
	server := sl.getServer(hostport)
	expire := time.Duration(sl.config.ExpireTime) * time.Second

	if server == nil {
		server = new(Server)
		server.timer = time.NewTimer(expire)
		server.payload = string(payload)

		if !sl.setServer(hostport, server) {
			return errors.New("Too many servers registered on host.")
		} else {
			log.Println(hostport, "added")
		}

		go func() {
			<-server.timer.C
			sl.deleteServer(hostport)
			log.Println(hostport, "expired")
		}()
	} else {
		if server.timer.Reset(expire) {
			// Server is still on list, just update payload.
			sl.setServerPayload(hostport, string(payload))
			log.Println(hostport, "updated")
		} else {
			// Server no longer on list, update payload and readd.
			server.payload = string(payload)
			sl.setServer(hostport, server)
			log.Println(hostport, "readded")
		}
	}

	return nil
}

func (sl *ServerList) getServer(hostport string) *Server {
	sl.lock.RLock()
	defer sl.lock.RUnlock()
	return sl.servers[hostport]
}

func (sl *ServerList) setServer(hostport string, server *Server) bool {
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		log.Fatal(err)
	}

	sl.lock.Lock()
	defer sl.lock.Unlock()

	if sl.serverCount[host] == sl.config.MaxServersPerHost {
		return false
	} else {
		sl.serverCount[host]++
	}

	sl.servers[hostport] = server
	return true
}

func (sl *ServerList) deleteServer(hostport string) {
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		log.Fatal(err)
	}

	sl.lock.Lock()
	{
		delete(sl.servers, hostport)

		sl.serverCount[host]--
		if sl.serverCount[host] == 0 {
			delete(sl.serverCount, host)
		}
	}
	sl.lock.Unlock()
}

func (sl *ServerList) setServerPayload(hostport string, payload string) {
	sl.lock.Lock()
	sl.servers[hostport].payload = payload
	sl.lock.Unlock()
}
