package util

import (
	"net"
	"strings"
	"unicode"
)

func IsValidIpAddress(addr string) bool {
	if addr == "" {
		return false
	}

	ip := net.ParseIP(addr)
	if ip != nil {
		return true
	}

	// web host
	pc := strings.Count(addr, ".")
	if pc < 1 || pc > 2 {
		return false
	}

	index := strings.Index(addr, ".")
	if index == 0 {
		return false
	}
	index = strings.LastIndex(addr, ".")
	end := addr[index+1:]
	if end == "" {
		return false
	}

	for _, t := range end {
		if !unicode.IsLetter(t) {
			return false
		}
	}
	return true
}
