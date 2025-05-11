package network

import (
	"fmt"
	"testing"
)

func TestLocal(t *testing.T) {
	ip := GetLocalIP()
	fmt.Printf("Local IP: %s\n", ip)
	name := GetLocalHostName()
	fmt.Printf("Local hostname: %s\n", name)
	mac := GetMACAddress()
	fmt.Printf("Local MAC address: %s\n", mac)
}
