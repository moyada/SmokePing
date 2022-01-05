package ping

import (
	"fmt"
	"strings"
	"time"
)

func AuthProtocol(host string) (bool, error) {
	pinger, err := NewPinger(host)
	if err != nil {
		return false, err
	}

	pinger.Count = 1
	pinger.Timeout = 3 * time.Second
	err = pinger.Run()
	if err == nil {
		return pinger.Privileged(), nil
	}

	permitErr := strings.Contains(err.Error(), "operation not permitted")

	pinger.SetPrivileged(!pinger.Privileged())
	err = pinger.Run()

	if err == nil && permitErr {
		fmt.Println("[Operation Error] icmp socket not permitted, switch to udp.")
	}
	return pinger.Privileged(), err
}
