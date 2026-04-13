package port

import (
	"errors"
	"fmt"
	"net"

	"github.com/IsmailAki/devc/internal/state"
)

const (
	BasePort = 2222
	MaxPort  = 22999
)

var ErrNoAvailablePort = errors.New("no available port in range")

func Allocate() (int, error) {
	containers, err := state.ListContainers()
	if err != nil {
		return 0, err
	}

	usedPorts := make(map[int]bool)
	for _, name := range containers {
		containerState, err := state.LoadState(name)
		if err != nil {
			continue
		}
		usedPorts[containerState.SSHPort] = true
	}

	for port := BasePort; port <= MaxPort; port++ {
		if usedPorts[port] {
			continue
		}

		if isAvailable(port) {
			return port, nil
		}
	}

	return 0, ErrNoAvailablePort
}

func isAvailable(port int) bool {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return false
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return false
	}
	l.Close()
	return true
}
