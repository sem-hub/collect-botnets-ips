package utils

import (
	"bufio"
	"compress/gzip"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sem-hub/collect-botnets-ips/internal/app/configs"
)

func GetOldIpset(config *configs.Config) map[string]bool {
	res := make(map[string]bool)
	f, err := os.Open(config.IpSetFile)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimLeft(scanner.Text(), "add dropips ")
		if net.ParseIP(line).To4() != nil {
			res[line] = true
		}
	}
	f.Close()
	return res
}

func GetNewAbuseDB(config *configs.Config) map[string]bool {
	res := make(map[string]bool)

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.abuseipdb.com/api/v2/blacklist", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Accept", "text/plain")
	req.Header.Add("Key", config.AbuseipdbKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		ip := scanner.Text()
		if net.ParseIP(ip).To4() != nil {
			res[ip] = true
		}
	}

	return res
}

func GetBanUnban(config *configs.Config) map[string]bool {
	res := make(map[string]bool)

	files, err := filepath.Glob(config.Fail2banLogs)
	if err != nil {
		log.Fatal()
	}
	for _, fn := range files {
		f, err := os.Open(fn)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		var r io.Reader
		if fn[len(fn)-3:] == ".gz" {
			r, err = gzip.NewReader(f)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			r = f
		}
		bufferedContents := bufio.NewReader(r)

		scanner := bufio.NewScanner(bufferedContents)
		for scanner.Scan() {
			line := scanner.Text()
			if (strings.Contains(line, "Ban") || strings.Contains(line, "Unban")) && !strings.Contains(line, "Increase") && !strings.Contains(line, "Restore") {
				space := regexp.MustCompile(`\s+`)
				fields := strings.Split(space.ReplaceAllString(line, " "), " ")
				if len(fields) != 8 {
					log.Printf("Parse error line: %s\n", line)
				} else {
					ip := fields[7]
					if net.ParseIP(ip).To4() != nil {
						res[ip] = true
					}
				}
			}
		}
	}
	return res
}

func GetNeverIps(config *configs.Config) map[string]bool {
	res := make(map[string]bool)
	f, err := os.Open(config.NeverIpsFile)
	if err != nil {
		log.Print(err)
		return res
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if net.ParseIP(line).To4() != nil {
			res[line] = true
		}
	}

	return res
}
