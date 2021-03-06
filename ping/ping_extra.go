package ping

import (
	"fmt"
	"strings"
	"time"
)

func SelectSocket(host string) (bool, error) {
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
		fmt.Println("[Operation Warning] icmp socket is not permitted, select udp socket.")
	}
	return pinger.Privileged(), err
}
