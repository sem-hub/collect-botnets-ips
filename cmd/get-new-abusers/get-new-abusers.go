package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/sem-hub/collect-botnets-ips/internal/app/configs"
	"github.com/sem-hub/collect-botnets-ips/internal/app/utils"
)

var (
	configPath   string
	config       *configs.Config
	debug        bool
	need_abusedb bool
	verbose      bool
)

func init() {
	flag.StringVar(&configPath, "config-path", "configs/botnets-ipset.toml", "path to config file")
	flag.BoolVar(&need_abusedb, "abusedb", false, "Fetch abuse DB")
	flag.BoolVar(&verbose, "v", false, "Be verbose")
	flag.BoolVar(&debug, "d", false, "Turn on debug output")
}

func main() {
	flag.Parse()

	config = configs.NewConfig()
	_, err := toml.DecodeFile(configPath, config)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile("/var/log/get-new-ipset.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if !debug {
		log.SetOutput(f)
	}

	res := make(map[string]bool)
	if need_abusedb {
		if verbose {
			log.Println("Get Abuse DB:")
		}
		abusedb := utils.GetNewAbuseDB(config)

		/*
			f, err := os.Create("abuseip.db")
			if err != nil {
				log.Fatal(err)
			}
		*/
		for key := range abusedb {
			res[key] = true
			//	f.WriteString(key + "\n")
		}
		//f.Close()
	} else {
		banUnban := utils.GetBanUnban(config)
		for key := range banUnban {
			res[key] = true
		}

	}
	oldIpset := utils.GetOldIpset(config)
	newIpset := make(map[string]bool)
	neverIps := utils.GetNeverIps(config)
	for key := range res {
		if neverIps[key] {
			log.Printf("found never IP: %s\n", key)
		}
		if !oldIpset[key] && !neverIps[key] {
			newIpset[key] = true
		}
	}

	if verbose {
		if len(newIpset) > 0 {
			log.Printf("found %d new IPs\n", len(newIpset))
		} else {
			log.Println("nothing new found")
		}
	}

	i := 0
	for key := range newIpset {
		i++
		log.Printf("%s: add %d: %s\n", time.Now().String(), i, key)
		for _, server := range config.ServerAddr {
			utils.SendToServer(server, config.Token, "add", key, need_abusedb)
		}
	}
}
