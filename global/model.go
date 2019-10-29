package global

import "github.com/nsqio/go-nsq"

type ServerConfig struct {
	ServerId  int      `toml:"serverId"`
	Addr      string   `toml:"addr"`
	HttpPort  int      `toml:"http-port"`
	TcpPort   int      `toml:"tcp-port"`
}

type RedisConfig struct {
	Name  string  `toml:"server_name"`
	Addr  string  `toml:"addr"`
	Pwd   string  `toml:"password"`
}

type NsqConsumerConfig struct {
	Name    string  `toml:"nsq_name"`
	Addr    string  `toml:"addr"`
	Topic   string  `toml:"topic"`
	Channel string  `toml:"channel"`
}

type NsqProducerConfig struct {
	Name    string  `toml:"nsq_name"`
	Addr    string  `toml:"addr"`
	Topic   string  `toml:"topic"`
}

type ConsulConfig struct {
	Name         string   `toml:"server_name"`
	Addr         string   `toml:"consul_addr"`
	TimeOut      string   `toml:"time_out"`
	Interval     string   `toml:"interval"`
	Delete       string   `toml:"delete"`
}

type ClientConfig struct {
	Name        string   `toml:"service_name"`
	Consul      bool     `toml:"use_consul"`
	EndPoints   string   `toml:"endpoints"`
	ReadTimeout int      `toml:"read_timeout"`
}

type SqlGroupConfig struct {
	Name     string   `toml:"name"`
	Master   string   `toml:"master"`
	Slaves   []string `toml:"slaves"`
}

type Config struct {
	Server      ServerConfig        `toml:"server"`
	Consul      *ConsulConfig       `toml:"consul"`
	Redis       []RedisConfig       `toml:"redis"`
	Clients     []ClientConfig      `toml:"client"`
	NsqConsumer []NsqConsumerConfig `toml:"nsq_consumer"`
	NsqProducer []NsqProducerConfig `toml:"nsq_producer"`
	SqlGroups   []SqlGroupConfig    `toml:"database"`
}

type FreyaNsqHandler struct {
	Name      string
	consumer  *nsq.Consumer
	nsq.Handler
}
