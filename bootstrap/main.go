package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	defaultTTL     = 300                       // Default TTL value (300 seconds)
	configPath     = "/etc/config/config.json" // Path to the configuration file
	apiTokenEnvVar = "CLOUDFLARE_API_TOKEN"    // Environment variable for Cloudflare API token
)

// Config represents the structure of the configuration file
type Config struct {
	Zones       []Zone `json:"zones"`
	IPv4Enabled bool   `json:"ipv4_enabled"`
	IPv6Enabled bool   `json:"ipv6_enabled"`
	TTL         int    `json:"ttl"`
}

// Zone represents a DNS zone configuration
type Zone struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	UpdateRootDomain bool     `json:"update_root_domain"`
	Subdomains       []string `json:"subdomains"`
}

// DNSRecord represents a DNS record
type DNSRecord struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

var config Config // Holds loaded configuration data

// main is the entry point of the program; it loads configuration, handles signals, and runs IP updates based on CLI flags.
func main() {
	if err := loadConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	handleSignals()

	if len(os.Args) > 1 && os.Args[1] == "--repeat" {
		for {
			setIPsOnRecords(getIPs())
			time.Sleep(time.Duration(config.TTL) * time.Second)
		}
	} else {
		setIPsOnRecords(getIPs())
	}
}

// loadConfig reads and parses the configuration file from disk into the config variable.
func loadConfig() error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading config.json: %w", err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("error parsing config.json: %w", err)
	}
	if config.TTL < 60 {
		config.TTL = defaultTTL
		log.Printf("TTL is too low - defaulting to %d seconds", defaultTTL)
	}
	return nil
}

// handleSignals sets up signal handling for graceful shutdown on system interrupt or termination.
func handleSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for range c {
			log.Println("Stopping main thread...")
			os.Exit(0)
		}
	}()
}

// getIPs retrieves the public IPv4 and IPv6 addresses, if enabled in the config, and returns them as DNS records.
func getIPs() map[string]DNSRecord {
	ips := make(map[string]DNSRecord)
	if config.IPv4Enabled {
		ipv4, err := fetchNetworkTrace("https://1.1.1.1/cdn-cgi/trace", "ip")
		if err != nil {
			ipv4, err = fetchNetworkTrace("https://1.0.0.1/cdn-cgi/trace", "ip")
			if err != nil {
				log.Println("IPv4 not detected.")
			}
		}
		if ipv4 != "" {
			ips["ipv4"] = DNSRecord{Type: "A", Content: ipv4, TTL: config.TTL}
		}
	}
	if config.IPv6Enabled {
		ipv6, err := fetchNetworkTrace("https://[2606:4700:4700::1111]/cdn-cgi/trace", "ip")
		if err != nil {
			ipv6, err = fetchNetworkTrace("https://[2606:4700:4700::1001]/cdn-cgi/trace", "ip")
			if err != nil {
				log.Println("IPv6 not detected.")
			}
		}
		if ipv6 != "" {
			ips["ipv6"] = DNSRecord{Type: "AAAA", Content: ipv6, TTL: config.TTL}
		}
	}
	return ips
}

// fetchNetworkTrace fetches the value of a specified key from the response of a network trace URL.
func fetchNetworkTrace(url, key string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(line, key+"=") {
			return strings.Split(line, "=")[1], nil
		}
	}
	return "", fmt.Errorf("%s not found", key)
}

// setIPsOnRecords updates DNS records with the current IPs by iterating over each IP address record.
func setIPsOnRecords(ips map[string]DNSRecord) {
	for _, ip := range ips {
		updateRecords(ip)
	}
}

// updateRecords updates DNS records for all configured zones and subdomains based on the provided DNS record.
func updateRecords(record DNSRecord) {
	for _, zone := range config.Zones {
		endpoint := fmt.Sprintf("zones/%s/dns_records?per_page=5000000&type=%s", zone.ID, record.Type)
		existingRecords, err := makeCloudflareAPIRequest(endpoint, "GET", nil)
		if err != nil {
			log.Printf("Error fetching existing DNS records: %v", err)
			continue
		}

		if zone.UpdateRootDomain {
			record.Name = zone.Name
			updateRecord(zone, existingRecords, record)
		}

		for _, subdomain := range zone.Subdomains {
			record.Name = fmt.Sprintf("%s.%s", subdomain, zone.Name)
			updateRecord(zone, existingRecords, record)
		}
	}
}

// updateRecord checks if a DNS record exists and needs updating; if so, it updates the record.
func updateRecord(zone Zone, existingRecords map[string]interface{}, record DNSRecord) {
	existingID, modified := findExistingRecord(existingRecords, record.Name, record.Content)

	if existingID != "" && modified {
		log.Printf("Updating record %v", record)
		endpoint := fmt.Sprintf("zones/%s/dns_records/%s", zone.ID, existingID)
		if _, err := makeCloudflareAPIRequest(endpoint, "PATCH", record); err != nil {
			log.Printf("Error updating record %v: %v", record, err)
		}
	} else if existingID == "" {
		log.Printf("Record %s doesn't exist", record.Name)
	}
}

// findExistingRecord checks if a DNS record with the specified FQDN exists and if its content differs from the desired content.
func findExistingRecord(existingRecords map[string]interface{}, fqdn, content string) (string, bool) {
	var existingID string
	modified := false

	if existingRecords != nil {
		for _, record := range existingRecords["result"].([]interface{}) {
			r := record.(map[string]interface{})
			if r["name"].(string) == fqdn {
				existingID = r["id"].(string)
				if r["content"].(string) != content {
					modified = true
				}
				break
			}
		}
	}
	return existingID, modified
}

// makeCloudflareAPIRequest performs an HTTP request to the Cloudflare API and returns the JSON-decoded response.
func makeCloudflareAPIRequest(endpoint, method string, data interface{}) (map[string]interface{}, error) {
	client := &http.Client{}
	url := "https://api.cloudflare.com/client/v4/" + endpoint

	var req *http.Request
	var err error

	if data != nil {
		body, _ := json.Marshal(data)
		req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	apiToken := os.Getenv(apiTokenEnvVar)
	if apiToken == "" {
		return nil, fmt.Errorf("API token environment variable %s is not set", apiTokenEnvVar)
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if result["errors"] != nil {
		return nil, fmt.Errorf("%v", result["errors"])
	}

	return result, nil
}
