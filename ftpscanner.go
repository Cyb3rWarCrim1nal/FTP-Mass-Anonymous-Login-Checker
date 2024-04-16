package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/jlaffaye/ftp"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// GeoInfo represents ISP and country information.
type GeoInfo struct {
	ISP     string `json:"isp"`
	Country string `json:"country"`
}

// Result represents the result of an FTP connection attempt.
type Result struct {
	Host     string
	Success  bool
	ErrorMsg string
}

// Function to get geo-location information of an IP address
func getGeoInfo(ip string) (GeoInfo, error) {
	geoInfo := GeoInfo{}

	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)
	resp, err := http.Get(url)
	if err != nil {
		return geoInfo, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return geoInfo, fmt.Errorf("IP geolocation request failed with status: %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&geoInfo)
	if err != nil {
		return geoInfo, err
	}

	return geoInfo, nil
}

// Function to resolve a hostname to an IP address
func resolveIP(hostname string) (string, error) {
	addrs, err := net.LookupHost(hostname)
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ip := net.ParseIP(addr); ip != nil && ip.To4() != nil {
			return addr, nil
		}
	}

	return "", fmt.Errorf("no IPv4 address found for hostname %s", hostname)
}

// Function for anonymous login to FTP server
func anonymousLogin(host string, wg *sync.WaitGroup, results chan<- Result, foundFile *os.File) {
	defer wg.Done()

	ftpAddress := host + ":21"

	// Establish connection to FTP server
	client, err := ftp.Dial(ftpAddress, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		results <- Result{Host: host, Success: false, ErrorMsg: fmt.Sprintf("FTP Connection Error: %v", err)}
		return
	}
	defer client.Quit()

	// Anonymous login
	err = client.Login("anonymous", "")
	if err != nil {
		results <- Result{Host: host, Success: false, ErrorMsg: fmt.Sprintf("FTP Anonymous Login Failed: %v", err)}
		return
	}

	// If login successful
	results <- Result{Host: host, Success: true, ErrorMsg: ""}

	// Resolve IP address only if login successful
	ip := host
	if net.ParseIP(host) == nil {
		// Resolve IP address for hostname
		resolvedIP, err := resolveIP(host)
		if err != nil {
			fmt.Printf("Failed to resolve hostname %s: %v\n", host, err)
			return
		}
		ip = resolvedIP
	}

	// Get country and ISP information for resolved IP address
	geoInfo, err := getGeoInfo(ip)
	if err != nil {
		fmt.Printf("Failed to get geo information for IP %s: %v\n", ip, err)
		return
	}

	// Write result to found file
	if foundFile != nil {
		_, err := foundFile.WriteString(fmt.Sprintf("%s, ISP: %s, Country: %s\n", host, geoInfo.ISP, geoInfo.Country))
		if err != nil {
			fmt.Printf("Failed to write to found file: %v\n", err)
		}
	}
}

func main() {
	filePath := "hosts.txt"
	foundFilePath := "found.txt"

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error opening %s file: %v\n", filePath, err)
	}
	defer file.Close()

	foundFile, err := os.OpenFile(foundFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening %s file: %v\n", foundFilePath, err)
	}
	defer foundFile.Close()

	scanner := bufio.NewScanner(file)

	var wg sync.WaitGroup
	results := make(chan Result)

	numWorkers := 10
	semaphore := make(chan struct{}, numWorkers)

	for scanner.Scan() {
		host := strings.TrimSpace(scanner.Text())
		if host != "" {
			wg.Add(1)
			semaphore <- struct{}{}

			go func(hostname string) {
				defer func() {
					<-semaphore
				}()

				anonymousLogin(hostname, &wg, results, foundFile)
			}(host)
		}
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		if result.Success {
			// Handle successful login
			fmt.Printf("[+] %s FTP Anonymous Login Succeeded.\n", result.Host)
			// Display location information only for successful connections
			geoInfo, err := getGeoInfo(result.Host)
			if err != nil {
				fmt.Printf("Failed to get geo information for IP %s: %v\n", result.Host, err)
				continue
			}
			fmt.Printf("    ISP: %s\n", geoInfo.ISP)
			fmt.Printf("    Country: %s\n", geoInfo.Country)
		} else {
			// Handle failed login
			fmt.Printf("[-] %s FTP Anonymous Login Failed: %v\n", result.Host, result.ErrorMsg)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading %s file: %v\n", filePath, err)
	}
}