package network

import (
	"bytes"
	"encoding/binary"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	HeaderXRealIP       = "X-Real-Ip"
	HeaderXForwardedFor = "X-Forwarded-For"
)

func GetClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	ra := r.RemoteAddr
	if ip := r.Header.Get(HeaderXForwardedFor); ip != "" {
		ra = strings.TrimSpace(strings.Split(ip, ",")[0])
	} else if ip := r.Header.Get(HeaderXRealIP); ip != "" {
		ra = strings.TrimSpace(strings.Split(ip, ",")[0])
	} else if strings.Contains(ra, ":") {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}

// GetLocalHostName returns the local hostname or the local IP if the hostname cannot be resolved
func GetLocalHostName() string {
	name, err := os.Hostname()
	if err != nil {
		name = GetLocalIP()
	}
	return name
}

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {

		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// GetMACAddress returns the local mac address
func GetMACAddress() (addr string) {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			if i.Flags&net.FlagUp != 0 && !bytes.Equal(i.HardwareAddr, nil) {
				// Don't use random as we have a real address
				addr = i.HardwareAddr.String()
				break
			}
		}
	}
	return
}

func IPInRange(ip string, list []string) bool {
	userIP := net.ParseIP(ip)
	if len(list) == 0 {
		return true
	}

	for _, against := range list {
		switch {
		case against == "*":
			return true
		case strings.Contains(against, "/"):
			_, in, e := net.ParseCIDR(against)
			if e != nil {
				continue
			}
			if in.Contains(userIP) {
				return true
			}
		case strings.Contains(against, "-"):
			ipRange := strings.Split(against, "-")
			start, _ := strconv.Atoi(strings.Split(ipRange[0], ".")[3])
			end, _ := strconv.Atoi(ipRange[1])
			between, _ := strconv.Atoi(strings.Split(ip, ".")[3])
			if between >= start && between <= end {
				return true
			}
		default:
			if net.ParseIP(against).Equal(userIP) {
				return true
			}
		}
	}
	return false
}

// IPToInt encodes a string representation of a IP to a uint32
func IPToInt(ip string) uint32 {
	i := net.ParseIP(ip)
	if len(i) == 16 {
		return binary.BigEndian.Uint32(i[12:16])
	}
	return binary.BigEndian.Uint32(i)
}

// IntToIP decodes a uint32 encoded IP to a string representation
func IntToIP(ip uint32) string {
	i := make(net.IP, 4)
	binary.BigEndian.PutUint32(i, ip)
	return i.String()
}

// CIDRMatch checks if specific ip belongs to specific cidr block (ip/mask)
func CIDRMatch(ip net.IP, cidr string) bool {
	_, cidrNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}

	return cidrNet.Contains(ip)
}
