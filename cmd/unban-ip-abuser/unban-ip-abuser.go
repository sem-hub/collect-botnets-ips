package main

import (
	"flag"
	"log"
	"net"

	"github.com/BurntSushi/toml"
	"github.com/sem-hub/collect-botnets-ips/internal/app/configs"
	"github.com/sem-hub/collect-botnets-ips/internal/app/utils"
)

var (
	configPath string
	config     *configs.Config
	debug      bool
)

func init() {
	flag.StringVar(&configPath, "config-path", "configs/botnets-ipset.toml", "path to config file")
}

func unban(file string, addr string) {
	if net.ParseIP(addr).To4() == nil {
		log.Fatalln("Wrong IPv4:", addr)
	}

	log.Println("Send to servers")
	for _, server := range config.ServerAddr {
		log.Println((server))
		utils.SendToServer(server, config.Token, "unbanip", addr, false)
	}
}

func main() {
	flag.Parse()

	config = configs.NewConfig()
	_, err := toml.DecodeFile(configPath, config)
	if err != nil {
		log.Fatal(err)
	}

	if len(flag.Args()) == 0 {
		log.Fatal("Nothing to do")
	}

	for _, key := range flag.Args() {
		unban(config.NeverIpsFile, key)
	}
}
