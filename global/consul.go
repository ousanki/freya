package global

import (
	"fmt"
	"github.com/ousanki/freya/log"
	"time"

	"github.com/hashicorp/consul/api"
)

func newConsul(c *Config) *api.Client {
	config := api.DefaultConfig()
	// 设置consul地址
	config.Address = c.Consul.Addr
	// 创建client
	client, err := api.NewClient(config)
	if err != nil {
		panic(err.Error())
	}
	return client
}

func registerConsul(c *Config, client *api.Client) {
	// 注册自己的服务
	registration := &api.AgentServiceRegistration{
		ID:   fmt.Sprintf("%s-%d", c.Consul.Name, c.Server.ServerId),
		Name: c.Consul.Name,
		Port: c.Server.HttpPort,
		Tags: []string{
			fmt.Sprintf("Game:%d", (c.Server.ServerId - 100)/100 + 1),
			fmt.Sprintf("Server:%d", c.Server.ServerId),
		},
		Address: c.Consul.Addr,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d%s", c.Server.Addr, c.Server.HttpPort, "/check"),
			Timeout:                        c.Consul.TimeOut,
			Interval:                       c.Consul.Interval,
			DeregisterCriticalServiceAfter: c.Consul.Delete,
		},
	}
	err := client.Agent().ServiceRegister(registration)
	if err != nil {
		panic(err.Error())
	}
}

func updateClient(name string, g *Global) {
	if g.consul == nil {
		return
	}
	client, ext := g.clients[name]
	if !ext {
		return
	}
	if !client.UseConsul {
		return
	}
	for {
		services, metainfo, err := g.consul.Health().Service(name, "", true,
			&api.QueryOptions{
				WaitIndex: client.LastIndex,
				WaitTime:  time.Second * 30,
			})
		if err != nil {
			log.GetLogger().Warningf("Consul | updateClient error:%v", err)
			continue
		}
		client.LastIndex = metainfo.LastIndex
		var eps []string
		for _, service := range services {
			eps = append(eps, fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port))
		}
		client.Mu.Lock()
		client.ConsulEps = eps
		client.Mu.Unlock()

		time.Sleep(time.Second * 15)
	}
}
