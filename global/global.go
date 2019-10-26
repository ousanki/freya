package global

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/hashicorp/consul/api"
	"github.com/nsqio/go-nsq"
)

type NsqProducer struct {
	P     *nsq.Producer
	Topic string
}

func (np *NsqProducer) Publish(msg []byte) error {
	return np.P.Publish(np.Topic, msg)
}

func (np *NsqProducer) MultiPublish(msgs [][]byte) error {
	return np.P.MultiPublish(np.Topic, msgs)
}

type NsqConsumer struct {
	C    *nsq.Consumer
	Addr string
}

func (nc *NsqConsumer) Do(h nsq.Handler) error {
	nc.C.AddHandler(h)
	return nc.C.ConnectToNSQD(nc.Addr)
}

type RpcClient struct {
	Name        string
	Eps         []string
	ConsulEps   []string
	Mu          sync.RWMutex
	UseConsul   bool
	LastIndex   uint64
	ReadTimeout int
}

type Global struct {
	ServerId    int
	HttpPort    int
	TcpPort     int
	Addr        string
	consul      *api.Client
	clients     map[string]*RpcClient
	rds         map[string]*redis.Pool
	nsqConsumer map[string]*NsqConsumer
	nsqProducer map[string]*NsqProducer
}

func (g *Global) initRedis(name, addr, pwd string) {
	db := &redis.Pool{
		MaxIdle:     10000,
		MaxActive:   10000,
		IdleTimeout: 5 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				return nil, err
			}
			if pwd == "" {
				return c, nil
			}
			if _, err := c.Do("AUTH", pwd); err != nil {
				_ = c.Close()
				return nil, err
			}
			return c, nil
		},
	}
	g.rds[name] = db
}

func (g *Global) initNsqConsumer(name, topic, channel, addr string) {
	var err error
	// 初始化消费者
	cfg := nsq.NewConfig()
	cfg.MsgTimeout = time.Second * 5
	cfg.MaxAttempts = 65535
	cfg.MaxRequeueDelay = time.Hour
	cfg.HeartbeatInterval = time.Second * 30

	var c NsqConsumer
	c.C, err = nsq.NewConsumer(topic, channel, cfg)
	if err != nil {
		panic(err.Error())
	}
	c.Addr = addr
	g.nsqConsumer[name] = &c
}

func (g *Global) initNsqProducer(name, topic, addr string) {
	var err error

	cfg := nsq.NewConfig()
	cfg.MsgTimeout = time.Second * 5
	cfg.MaxAttempts = 65535
	cfg.MaxRequeueDelay = time.Hour

	var p NsqProducer
	p.P, err = nsq.NewProducer(addr, cfg)
	if nil != err {
		panic(err.Error())
	}

	if err := p.P.Ping(); err != nil {
		panic(err.Error())
	}
	p.Topic = topic

	g.nsqProducer[name] = &p
}

func (g *Global) initClient(name string, use bool, eps string, rto int) {
	var endPoints []string
	endPoints = strings.Split(eps, ",")
	client := &RpcClient{
		Name:        name,
		UseConsul:   use,
		ReadTimeout: rto,
		Eps:         endPoints,
	}
	g.clients[name] = client

	go updateClient(name)
}

func GetRedis(name string) redis.Conn {
	if db, has := G.rds[name]; has {
		return db.Get()
	}
	return nil
}

func GetClient(name string) *RpcClient {
	if c, has := G.clients[name]; has {
		return c
	}
	return nil
}

func getNsqConsumer(name string) *NsqConsumer {
	return G.nsqConsumer[name]
}

func GetNsqProducer(name string) *NsqProducer {
	return G.nsqProducer[name]
}

func RunNsqConsumers(handlers ...*FreyaNsqHandler) {
	for _, h := range handlers {
		c := getNsqConsumer(h.Name)
		if c == nil {
			panic(fmt.Sprintf("nsq consumer [%s], is nil", h.Name))
		}
		h.consumer = c.C
		err := c.Do(h)
		if err != nil {
			panic(err.Error())
		}
	}
}
