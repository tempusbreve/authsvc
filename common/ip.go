package common // import "breve.us/authsvc/common"

import (
	"net/http"
	"strings"
)

func queryIP(r *http.Request) string {
	vars := r.URL.Query()
	ip := vars.Get("ip")
	if ip == "" {
		ip = remoteIP(r)
	}
	return ip
}

func lastIP(r *http.Request) string {
	return cleanIP(r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")])
}

func remoteIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-Ip")
	if ip == "" {
		if f := forwarders(r); len(f) > 0 {
			// get FIRST forwarder
			ip = f[0]
		}
	}
	if ip == "" {
		ip = lastIP(r)
	}
	return cleanIP(ip)
}

func forwarders(r *http.Request) []string {
	var f []string
	if forwards, ok := r.Header["X-Forwarded-For"]; ok {
		for _, fw := range forwards {
			for _, s := range strings.Split(fw, ",") {
				if ip := cleanIP(s); ip != "" {
					f = append(f, ip)
				}
			}
		}
	}
	return f
}

func cleanIP(ip string) string {
	return strings.Map(
		func(r rune) rune {
			switch r {
			case
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
				'A', 'B', 'C', 'D', 'E', 'F',
				'a', 'b', 'c', 'd', 'e', 'f',
				':', '.', '/':
				return r
			default:
				return -1
			}
		}, ip)
}
