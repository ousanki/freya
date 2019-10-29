package global

import (
	"github.com/BurntSushi/toml"
	"github.com/garyburd/redigo/redis"
)

var G *Global

func newToml() *Config {
	var c Config
	_, err := toml.DecodeFile("./etc/config.toml", &c)
	if err != nil {
		panic(err.Error())
	}
	return &c
}

func newGlobal(c *Config) *Global {
	g := Global{
		HttpPort:    c.Server.HttpPort,
		TcpPort:     c.Server.TcpPort,
		ServerId:    c.Server.ServerId,
		Addr:        c.Server.Addr,
		rds:         make(map[string]*redis.Pool),
		clients:     make(map[string]*RpcClient),
		nsqConsumer: make(map[string]*NsqConsumer),
		nsqProducer: make(map[string]*NsqProducer),
		sqlGroups:   make(map[string]*Group),
	}
	for _, r := range c.Redis {
		g.initRedis(r.Name, r.Addr, r.Pwd)
	}
	for _, nc := range c.NsqConsumer {
		g.initNsqConsumer(nc.Name, nc.Topic, nc.Channel, nc.Addr)
	}
	for _, np := range c.NsqProducer {
		g.initNsqProducer(np.Name, np.Topic, np.Addr)
	}
	if c.Consul != nil {
		g.consul = newConsul(c)
		registerConsul(c, g.consul)
	}
	for _, c := range c.Clients {
		g.initClient(c.Name, c.Consul, c.EndPoints, c.ReadTimeout)
	}
	for _, group := range c.SqlGroups {
		g.initDatabase(group.Name, group.Master, group.Slaves)
	}

	return &g
}

func InitGlobal() {
	G = newGlobal(newToml())
}