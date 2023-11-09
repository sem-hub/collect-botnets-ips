package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
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

type apiHandler struct{}

func (apiHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

func sendList(w http.ResponseWriter, r *http.Request) {
	ipset := utils.GetOldIpset(config)

	resp := utils.JsonResponse{Status: "OK"}
	for ip, _ := range ipset {
		resp.Data = append(resp.Data, ip)
	}
	fmt.Fprintln(w, utils.SendAsJson(resp))
}

func findIp(w http.ResponseWriter, r *http.Request, ip string) {
	ipset := utils.GetOldIpset(config)

	resp := utils.JsonResponse{}
	if ipset[ip] {
		resp.Status = "Found"
		log.Print("Found")
	} else {
		resp.Status = "Not Found"
		log.Print("Not found")
	}

	fmt.Fprintln(w, utils.SendAsJson(resp))
}

func deleteIp(w http.ResponseWriter, r *http.Request, ip string) {
	neverIps := utils.GetNeverIps(config)
	if neverIps[ip] {
		log.Println("Already in Never file. Process unban anyway.")
	} //else {
	//	log.Println("Add " + ip + " in Never file.")
	//	err := utils.AppendFile(config.NeverIpsFile, ip)
	//	if err != nil {
	//		log.Print(err)
	//	}
	//}

	ipset := utils.GetOldIpset(config)

	if !ipset[ip] {
		resp := utils.JsonResponse{Status: "error", ErrMsg: "IP not Found"}
		fmt.Fprintln(w, utils.SendAsJson(resp))
		return
	}
	delete(ipset, ip)

	var content []string
	// Get first line (ipset config)
	f, err := os.Open(config.IpSetFile)
	if err != nil {
		resp := utils.JsonResponse{Status: "error", ErrMsg: "IpSet file open error"}
		fmt.Fprintln(w, utils.SendAsJson(resp))
		log.Printf("Can't open IpSet file: %s", err)
		return
	}

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		resp := utils.JsonResponse{Status: "error", ErrMsg: "IpSet file read error"}
		fmt.Fprintln(w, utils.SendAsJson(resp))
		log.Printf("IpSet file read error: %s", err)
		return
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
		resp := utils.JsonResponse{Status: "error", ErrMsg: "IpSet file write error"}
		fmt.Fprintln(w, utils.SendAsJson(resp))
		log.Printf("Write error: %s", err)
		return
	}

	out, err := utils.OsExec("/sbin/ipset", "del dropips "+ip)
	if err != nil {
		resp := utils.JsonResponse{Status: "error", ErrMsg: "IpSet exec error"}
		fmt.Fprintln(w, utils.SendAsJson(resp))
		log.Printf("Shell command error: %s. %s", err, out)
		return
	}
}

func addIp(w http.ResponseWriter, r *http.Request, ip string) {
	ipset := utils.GetOldIpset(config)

	if ipset[ip] {
		resp := utils.JsonResponse{Status: "error", ErrMsg: "IP already exists"}
		fmt.Fprintln(w, utils.SendAsJson(resp))
		return
	}
	err := utils.AppendFile(config.IpSetFile, "add dropips "+ip)
	if err != nil {
		resp := utils.JsonResponse{Status: "error", ErrMsg: "Append file error"}
		fmt.Fprintln(w, utils.SendAsJson(resp))
		log.Printf("APPEND file error: %s", err)
		return
	}
	out, err := utils.OsExec("/sbin/ipset", "add dropips "+ip)
	if err != nil {
		resp := utils.JsonResponse{Status: "error", ErrMsg: "IpSet exec error"}
		fmt.Fprintln(w, utils.SendAsJson(resp))
		log.Printf("Shell command error: %s. %s", err, out)
		return
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Token")
	if token != config.Token {
		resp := utils.JsonResponse{Status: "error", ErrMsg: "Not authorized"}
		fmt.Fprintln(w, utils.SendAsJson(resp))
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Unauthorizated acces from: %s. URI: %s", r.RemoteAddr, r.RequestURI)
		return
	}

	log.Printf("Method: %s", r.Method)
	log.Printf("URI: %s", r.URL.Path)

	if !strings.HasPrefix(r.URL.String(), "/api/v1/ips/") {
		resp := utils.JsonResponse{Status: "error", ErrMsg: "API error"}
		fmt.Fprintln(w, utils.SendAsJson(resp))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "API error")
		log.Print("API error")
		return
	}

	ip := strings.TrimPrefix(r.URL.String(), "/api/v1/ips/")
	log.Printf("IP: %s", ip)
	if ip != "" && net.ParseIP(ip).To4() == nil {
		resp := utils.JsonResponse{Status: "error", ErrMsg: "IP syntax error"}
		fmt.Fprintln(w, utils.SendAsJson(resp))
		w.WriteHeader(http.StatusOK)
		log.Print("IP syntax error")
		return
	}
	if r.Method == "GET" {
		if ip == "" {
			sendList(w, r)
		} else {
			findIp(w, r, ip)
		}
	} else if r.Method == "POST" {
		addIp(w, r, ip)
	} else if r.Method == "DELETE" {
		deleteIp(w, r, ip)
	} else {
		res := utils.JsonResponse{Status: "error", ErrMsg: "Unknown HTTP method"}
		fmt.Fprintln(w, utils.SendAsJson(res))
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

	// log both Stdout and file
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	config = configs.NewConfig()
	_, err = toml.DecodeFile(configPath, config)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Server run")
	mux := http.NewServeMux()
	mux.Handle("/api/", apiHandler{})
	mux.HandleFunc("/api/v1/ips/", handler)
	log.Fatal(http.ListenAndServeTLS(config.BindAddr, config.TlsCertFile, config.TlsKeyFile, logRequest(mux)))
}
