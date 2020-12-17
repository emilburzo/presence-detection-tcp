package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const EnvPort = "PORT"
const EnvHosts = "HOSTS"
const EnvHostsSeparator = "HOSTS_SEPARATOR"
const EnvWebhooks = "WEBHOOKS"
const EnvCheckDelayPresent = "CHECK_DELAY_PRESENT"
const EnvCheckDelayAbsent = "CHECK_DELAY_ABSENT"

const DefaultPort = "1234"
const DefaultHostsSeparator = ","
const DefaultCheckDelayPresent = "300" // don't kill the phone's battery
const DefaultCheckDelayAbsent = "30"   // no downside to being faster here

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
			time.Sleep(getDelayWhenPresent() * time.Second)
		} else {
			time.Sleep(getDelayWhenAbsent() * time.Second)
		}
	}
}

func isPresentOnNetwork(host string) bool {
	address := host + ":" + getEnv(EnvPort, DefaultPort)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		opErr := err.(*net.OpError).Err
		errType := fmt.Sprintf("%T", opErr)
		if errType == "*os.SyscallError" {
			// connection refused means the device is present
			return opErr.(*os.SyscallError).Err.(syscall.Errno) == syscall.ECONNREFUSED
		}

		// any other error (timeout, no route, ...) means the device is not connected to the network
		return false
	}

	// mobile phones shouldn't be listening on any ports, but just in case...
	_ = conn.Close()

	// connection succeeded -> present on network
	return true
}

func getHosts() []string {
	hosts := getEnv(EnvHosts, "")
	return strings.Split(hosts, getEnv(EnvHostsSeparator, DefaultHostsSeparator))
}

func getDelayWhenAbsent() time.Duration {
	return getDelay(EnvCheckDelayAbsent, DefaultCheckDelayAbsent)
}

func getDelayWhenPresent() time.Duration {
	return getDelay(EnvCheckDelayPresent, DefaultCheckDelayPresent)
}

func getDelay(env string, fallback string) time.Duration {
	delay, err := strconv.Atoi(getEnv(env, fallback))
	if err != nil {
		log.Fatal("invalid delay")
	}
	return time.Duration(delay)
}

func getEnv(key string, fallback string) string {
	value, present := os.LookupEnv(key)
	if !present {
		if fallback == "" {
			log.Fatalf("environment variable `%s` is not defined", key)
		} else {
			return fallback
		}
	}

	return value
}
