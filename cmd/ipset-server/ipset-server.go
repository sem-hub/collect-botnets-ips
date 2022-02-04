package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/sem-hub/collect-botnets-ips/internal/app/configs"
	"github.com/sem-hub/collect-botnets-ips/internal/app/utils"
)

var (
	configPath string
	config     *configs.Config
)

func unbanIpCmd(w http.ResponseWriter, r *http.Request, ip string) error {
	neverIps := utils.GetNeverIps(config)
	if neverIps[ip] {
		log.Println("Already in Never file. Process unban anyway.")
	} else {
		log.Println("Add " + ip + " in Never file.")
		err := utils.AppendFile(config.NeverIpsFile, ip)
		if err != nil {
			log.Print(err)
		}
	}

	ipset := utils.GetOldIpset(config)

	if !ipset[ip] {
		w.WriteHeader(http.StatusNotAcceptable)
		return errors.New("Not exists")
	}
	delete(ipset, ip)

	var content []string
	// Get first line (ipset config)
	f, err := os.Open(config.IpSetFile)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		log.Print("First line read error")
		return errors.New("Can't read")
	}
	content = append(content, scanner.Text())
	f.Close()
	log.Print("First line: " + content[0])
	for k := range ipset {
		content = append(content, "add dropips "+k)
	}

	log.Print("Write new " + config.IpSetFile)
	err = utils.RewriteFile(config.IpSetFile, content)
	if err != nil {
		return err
	}

	out, err := utils.OsExec("/sbin/ipset", "del dropips "+ip)
	if err != nil {
		log.Printf("Shell command error: %s. %s", err, out)
		return err
	}
	return nil
}

func addIpCmd(w http.ResponseWriter, r *http.Request, ip string) error {
	ipset := utils.GetOldIpset(config)

	if ipset[ip] {
		w.WriteHeader(http.StatusNotAcceptable)
		return errors.New("Already exists")
	}
	err := utils.AppendFile(config.IpSetFile, "add dropips "+ip)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("APPEND file error: %s", err)
		return err
	}
	out, err := utils.OsExec("/sbin/ipset", "add dropips "+ip)
	if err != nil {
		log.Printf("Shell command error: %s. %s", err, out)
		return err
	}

	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Token")
	if token != config.Token {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, http.StatusText(http.StatusUnauthorized))
		return
	}

	q := r.URL.Query()
	if len(q["cmd"]) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "CMD is not given")
		log.Print("CMD is not given")
		return
	}
	if len(q["ip"]) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "IP address is not given")
		log.Print("IP address is not given")
		return
	}
	cmd := q["cmd"][0]
	ip := q["ip"][0]
	if net.ParseIP(ip).To4() == nil {
		w.WriteHeader(http.StatusExpectationFailed)
		log.Print("IP syntax error")
		return
	}
	if cmd == "addip" {
		err := addIpCmd(w, r, ip)
		if err != nil {
			fmt.Fprintln(w, err)
			log.Print(err)
			return
		}
		//fmt.Fprintln(w, "Added")
	} else if cmd == "unbanip" {
		err := unbanIpCmd(w, r, ip)
		if err != nil {
			fmt.Fprintln(w, err)
			log.Print(err)
			return
		}
		//fmt.Fprintln(w, "Unbanned")
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, http.StatusText(http.StatusNotFound))
		return
	}
	w.WriteHeader(http.StatusOK)
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
	//log.Printf("Code: %d\n", status)
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws := &statusWriter{ResponseWriter: w, status: 200}
		handler.ServeHTTP(ws, r)
		log.Printf("%s %s \"%s\" %d\n", r.RemoteAddr[0:strings.LastIndex(r.RemoteAddr, ":")], r.Method, r.URL, ws.status)
	})
}

func init() {
	flag.StringVar(&configPath, "config-path", "configs/botnets-ipset.toml", "path to config file")
}

func main() {
	flag.Parse()
	if os.Getegid() != 0 {
		log.Fatal("Run as root please.")
	}
	f, err := os.OpenFile("/var/log/ipset-server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	config = configs.NewConfig()
	_, err = toml.DecodeFile(configPath, config)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Server run")
	http.HandleFunc("/api", handler)
	log.Fatal(http.ListenAndServeTLS(":8080", config.TlsCertFile, config.TlsKeyFile, logRequest(http.DefaultServeMux)))
}
