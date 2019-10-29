package global

import (
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"strconv"
	"time"
)

type Client struct {
	*gorm.DB
}

type Group struct {
	name string

	master  *Client
	replica []*Client
	next    uint64
	total   uint64
}

func parseConnAddress(address string) (string, int, int, int, error) {
	u, err := mysql.ParseDSN(address)
	if err != nil {
		return address, -1, -1, 0, err
	}
	q := u.Params
	idleQ, activeQ, lifetimeQ := q["max_idle"], q["max_active"], q["max_lifetime_sec"]
	maxIdle, _ := strconv.Atoi(idleQ)
	if maxIdle == 0 {
		maxIdle = 15
	}
	maxActive, _ := strconv.Atoi(activeQ)
	lifetime, _ := strconv.Atoi(lifetimeQ)
	if lifetime == 0 {
		lifetime = 1800
	}
	delete(q, "max_idle")
	delete(q, "max_active")
	delete(q, "max_lifetime_sec")
	return u.FormatDSN(), maxIdle, maxActive, lifetime, nil
}

func openDB(name, address string) (*Client, error) {
	addr, maxIdle, maxActive, lifetime, err := parseConnAddress(address)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open("mysql", addr)
	if err != nil {
		return nil, fmt.Errorf("open mysql [%s] master %s error %s", name, address, err)
	}
	db = db.Debug()
	db.DB().SetMaxIdleConns(maxIdle)
	db.DB().SetMaxOpenConns(maxActive)
	db.DB().SetConnMaxLifetime(time.Duration(lifetime) * time.Second)

	return &Client{DB: db}, err
}

func newGroup(name string, master string, slaves []string) (*Group, error) {
	g := Group{name: name}
	var err error
	g.master, err = openDB(name, master)
	if err != nil {
		return nil, err
	}
	g.replica = make([]*Client, 0, len(slaves))
	g.total = 0
	for _, slave := range slaves {
		c, err := openDB(name, slave)
		if err != nil {
			return nil, err
		}
		g.replica = append(g.replica, c)
		g.total++

	}
	return &g, nil
}
