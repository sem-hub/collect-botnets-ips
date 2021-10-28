package utils

import (
	"crypto/tls"
	"io/ioutil"
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
	req, err := http.NewRequest("GET", "https://"+server+"/api?cmd="+cmd+"&ip="+ip+addStr, nil)
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
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Server: %s, Read body error: %s", server, err)
	} /*else {
		log.Printf("Server %s: %s", server, text)
	}*/
	resp.Body.Close()
}
