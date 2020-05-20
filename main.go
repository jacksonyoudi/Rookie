package main

import (
	"git.garena.com/changyou.liang/entry/Server"
	"git.garena.com/changyou.liang/entry/conf"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	data, err := ioutil.ReadFile("./conf.yml")
	if err != nil {
		log.Println(err)
		return
	}

	cfg := new(conf.ServiceCfg)
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		log.Println(err)
		return
	}

	// verification cfg
	if len(cfg.Address) == 0 {
		log.Println("config no address!")
		return
	}

	// verification cfg
	if len(cfg.Redis.Address) == 0 {
		log.Println("config no redis")
		return
	}

	if len(cfg.MySql.Dsn) == 0 || len(cfg.MySql.Driver) == 0 {
		log.Println("config no dsn")
		return
	}

	// tls
	if len(cfg.Tls.Server.Pem) == 0 || len(cfg.Tls.Server.Key) == 0 {
		log.Println("conf no tls server")
		return
	}

	if len(cfg.Tls.Client.Pem) == 0 || len(cfg.Tls.Client.Key) == 0 {
		log.Println("conf no tls server")
		return
	}

	var server Server.Service
	server, err = Server.NewTcpService(cfg)
	if err != nil {
		log.Println(err)
		return
	}
	server.Start()

}
