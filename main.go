package main

import (
	"bytes"
	"encoding/json"
	"github.com/cloudflare/cloudflare-go"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/net/context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func getOldIp() string {
	return os.Getenv("old_ip")
}

type IP struct {
	Ip string `json:"ip"`
}

func getNewIp() string {
	req, err := http.Get("https://api.ipify.org/?format=json")
	if err != nil {
		return err.Error()
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(req.Body)

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err.Error()
	}
	var ip IP
	_ = json.Unmarshal(body, &ip)
	return ip.Ip
}

func fileChange(oldIp string, newIp string) {
	input, _ := ioutil.ReadFile(".env")
	output := bytes.Replace(input, []byte(oldIp), []byte(newIp), 1)
	_ = ioutil.WriteFile(".env", output, 0666)
}

func dnsChange(ip string) {
	// 使用你的CLOUDFLARE_API_TOKEN
	api, _ := cloudflare.NewWithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN"))
	// 拿取區域識別碼
	id, _ := api.ZoneIDByName(os.Getenv("DNS_NAME"))
	// Most API calls require a Context
	ctx := context.Background()
	// 拿取該區域的所有 DNS 記錄
	zone, _ := api.DNSRecords(ctx, id, cloudflare.DNSRecord{Type: "A", Name: os.Getenv("DNS_NAME")})

	// Update the record's value
	_ = api.UpdateDNSRecord(ctx, id, zone[0].ID, cloudflare.DNSRecord{Content: ip})

}

func main() {
	oldIp := getOldIp()
	newIp := getNewIp()
	if oldIp != newIp {
		fileChange(oldIp, newIp)
		dnsChange(newIp)
	}
}
