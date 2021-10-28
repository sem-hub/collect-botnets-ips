package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/sem-hub/collect-botnets-ips/internal/app/configs"
	"github.com/sem-hub/collect-botnets-ips/internal/app/utils"
)

var (
	configPath string
	config     *configs.Config
)

func addIpCmd(w http.ResponseWriter, r *http.Request, ip string) error {
	ipset := utils.GetOldIpset(config)

	if net.ParseIP(ip).To4() == nil {
		w.WriteHeader(http.StatusExpectationFailed)
		return errors.New("IP syntax error")
	}
	if ipset[ip] {
		w.WriteHeader(http.StatusNotAcceptable)
		return errors.New("Already exists")
	}
	f, err := os.OpenFile(config.IpSetFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("APPEND file error: %s", err)
		return err
	}
	fmt.Fprintln(f, "add dropips "+ip)
	f.Close()
	out, err := exec.Command("/sbin/ipset", "add", "dropips", ip).CombinedOutput()
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
	if cmd == "addip" {
		err := addIpCmd(w, r, ip)
		if err != nil {
			fmt.Fprintln(w, err)
			log.Print(err)
			return
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, http.StatusText(http.StatusNotFound))
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Added")
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
