package main

import (
	"log"
	"net"
	"os"
	"strings"
	"syscall"
	"time"
)

const EnvPort = "PORT"
const EnvHosts = "HOSTS"
const EnvWebhooks = "WEBHOOKS"

const DelayWhenPresentSeconds = 5
const DelayWhenAbsentSeconds = 5

const StatusPresent = "present"
const StatusAbsent = "absent"

var currentStatus = "unknown"

func main() {

	for {
		var newStatus string
		for _, host := range getHosts() {
			log.Printf("trying...")
			if isPresentOnNetwork(host) {
				newStatus = StatusPresent
				break
			} else {
				newStatus = StatusAbsent
			}
		}

		if currentStatus != newStatus {
			log.Printf("changing presence from `%s` to `%s`", currentStatus, newStatus)
			// todo notify websockets
			currentStatus = newStatus
		}

		if currentStatus == StatusPresent {
			time.Sleep(DelayWhenPresentSeconds * time.Second)
		} else {
			time.Sleep(DelayWhenAbsentSeconds * time.Second)
		}
	}
}

func isPresentOnNetwork(host string) bool {
	address := host + ":" + getEnv(EnvPort)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		// connection refused means the device is present
		// any other error (timeout, no route, ...) means the device is not connected to the network
		return err.(*net.OpError).Err.(*os.SyscallError).Err.(syscall.Errno) == syscall.ECONNREFUSED
	}

	// mobile phones shouldn't be listening on port 80, but just in case...
	_ = conn.Close()

	// connection succeeded -> present on network
	return true
}

func getHosts() []string {
	hosts := getEnv(EnvHosts)
	return strings.Split(hosts, ",")
}

func getEnv(key string) string {
	value, present := os.LookupEnv(key)
	if !present {
		log.Fatalf("environment variable `%s` is not defined", key)
	}

	return value
}
