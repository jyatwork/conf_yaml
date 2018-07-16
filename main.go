package main

import (
	"conf_yaml/conf"
	"fmt"
)

func initConfigs() {
	conf.Load("conf.yaml")

}

func main() {
	initConfigs()
	fmt.Println(conf.String("rabbitmq.queue.key"))
}
