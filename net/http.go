package net

import (
	stdcontext "context"
	"errors"
	"fmt"
	"freya/global"
	"freya/handler"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var httpServer *iris.Application
var Done chan bool

func init() {
	Done = make(chan bool)

	httpServer = iris.New()
	httpServer.Configure(iris.WithConfiguration(iris.Configuration{
		Charset: "UTF-8",
		DisablePathCorrection: true,
	}))
	iris.RegisterOnInterrupt(
		func() {
			ctx, _ := stdcontext.WithTimeout(stdcontext.Background(), 5*time.Second)
			_ = httpServer.Shutdown(ctx)
			Done <- true
		},
	)
}

func RegHttpHandlers(handlers ...handler.HttpHandler) {
	for _, h := range handlers {
		switch strings.ToUpper(h.Method) {
		case "POST":
			httpServer.Post(h.Path, h.Handlers...)
		case "GET":
			httpServer.Get(h.Path, h.Handlers...)
		case "DELETE":
			httpServer.Delete(h.Path, h.Handlers...)
		case "PUT":
			httpServer.Put(h.Path, h.Handlers...)
		case "ANY":
			httpServer.Any(h.Path, h.Handlers...)
		}
	}
}

func RegHttpHandlersParty(party iris.Party, handlers ...handler.HttpHandler) {
	for _, h := range handlers {
		switch strings.ToUpper(h.Method) {
		case "POST":
			party.Post(h.Path, h.Handlers...)
		case "GET":
			party.Get(h.Path, h.Handlers...)
		case "DELETE":
			party.Delete(h.Path, h.Handlers...)
		case "PUT":
			party.Put(h.Path, h.Handlers...)
		case "ANY":
			party.Any(h.Path, h.Handlers...)
		}
	}
}

func NewHttpParty(h handler.HttpHandler) iris.Party {
	return httpServer.Party(h.Path, h.Handlers...)
}

func Use(handlers ...context.Handler) {
	httpServer.Use(handlers...)
}

func StartHttpServer() {
	if global.G.HttpPort == 0 {
		return
	}
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", global.G.HttpPort),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1024 * 1024,
	}
	httpServer.Get("/check", healthCheck)
	go httpServer.Run(iris.Server(server))
}

func healthCheck(ctx iris.Context) {
	ctx.StatusCode(iris.StatusOK)
	ctx.WriteString("OK")
}

func PostHttp(service, api string, body io.Reader) ([]byte, error) {
	endpoint, err := getEndPoint(service)
	if err != nil {
		return nil, err
	}
	uri := fmt.Sprintf("http://%s%s", endpoint, api)
	rsp, err := http.Post(uri, "application/json", body)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	return ioutil.ReadAll(rsp.Body)
}

func GetHttp(service, api string, values url.Values) ([]byte, error) {
	endpoint, err := getEndPoint(service)
	if err != nil {
		return nil, err
	}
	uri := fmt.Sprintf("http://%s%s?%s", endpoint, api, values.Encode())
	rsp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	return ioutil.ReadAll(rsp.Body)
}

func getEndPoint(service string) (string, error) {
	c := global.GetClient(service)
	if c == nil {
		return "", errors.New(fmt.Sprintf("Not find client:%s", service))
	}
	if !c.UseConsul && len(c.ConsulEps) == 0 && len(c.Eps) == 0 {
		return "", errors.New(fmt.Sprintf("No endpoints, client:%s ", service))
	}
	var endpoint string
	if !c.UseConsul || len(c.ConsulEps) == 0 {
		// use eps
		ran := rand.Intn(len(c.Eps))
		endpoint = c.Eps[ran]
	} else {
		// use consul eps
		c.Mu.RUnlock()
		ran := rand.Intn(len(c.ConsulEps))
		endpoint = c.ConsulEps[ran]
		c.Mu.RUnlock()
	}
	return endpoint,  nil
}