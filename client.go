package main

import (
	"fmt"
	"git.garena.com/changyou.liang/entry/Client"
	"log"
	"time"
)

func main() {
	client, err := Client.NewClient("changyou", "localhost:50000", "/Users/changyouliang/project/goproject/entrytask/cert/client.pem", "/Users/changyouliang/project/goproject/entrytask/cert/client.key", "/Users/changyouliang/project/goproject/entrytask/cert/server.pem")
	if err != nil {
		log.Println(err)
		return
	}
	cookie, err := client.Login("shopee")
	if err != nil {
		log.Println(err)
		return
	}
	client.Conn.Close()

	fmt.Println("write message")

	for i := 1; i < 100; i++ {
		client, err := Client.NewClient("changyou", "localhost:50000", "/Users/changyouliang/project/goproject/entrytask/cert/client.pem", "/Users/changyouliang/project/goproject/entrytask/cert/client.key", "/Users/changyouliang/project/goproject/entrytask/cert/server.pem")
		if err != nil {
			log.Println(err)
			return
		}
		client.SetCookie(cookie)
		fmt.Println(i)
		err = client.WriteSecureMessage("liangchangyou"+string(i), "hello word"+string(i))
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second * 1)
		}
		client.Conn.Close()
	}

	for i := 1; i < 1000; i++ {
		client, err := Client.NewClient("changyou", "localhost:50000", "/Users/changyouliang/project/goproject/entrytask/cert/client.pem", "/Users/changyouliang/project/goproject/entrytask/cert/client.key", "/Users/changyouliang/project/goproject/entrytask/cert/server.pem")
		if err != nil {
			log.Println(err)
			return
		}
		client.SetCookie(cookie)

		message, err := client.ReadSecureMessage("liangchangyou" + string(i))
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second * 1)
		}
		log.Println(message)
		client.Conn.Close()
	}

}
