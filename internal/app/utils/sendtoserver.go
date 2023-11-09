package utils

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

func SendToServer(server string, token string, cmd string, ip string, abusedb bool) {
	// Ignore self signed certificate error
	// Set a ServerName field to prevent name checking errors
	serverName := strings.Split(server, ":")[0]
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true,
			ServerName: serverName},
	}
	client := &http.Client{Transport: tr}
	addStr := ""
	if abusedb {
		addStr = "&abusedb"
	}
	var req *http.Request
	var err error
	if cmd == "add" {
		req, err = http.NewRequest("POST", "https://"+server+"/api/v1/ips/"+ip+addStr, nil)
	} else if cmd == "remove" {
		req, err = http.NewRequest("DELETE", "https://"+server+"/api/v1/ips/"+ip+addStr, nil)
	}
	if err != nil {
		log.Fatal(err)
	}
	// Do not keep-alive
	req.Close = true
	req.Header.Add("Token", token)
	// The same name as above
	req.Header.Add("Host", serverName)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		log.Printf("Server: %s: %s Code: %d", server, ip, resp.StatusCode)
	}
	jsonText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Server: %s, Read body error: %s", server, err)
	} else {
		respStruct := JsonResponse{}
		err := json.Unmarshal(jsonText, respStruct)
		if err != nil {
			log.Printf("Server response is ill: %s", jsonText)
		} else {
			log.Printf("Server %s: response status: %s", server, respStruct.Status)
		}
	}
	resp.Body.Close()
}
