package switchboard

import (
	"errors"
	"fmt"
	"net"

	"github.com/pivotal-golang/lager"
)

type Backend interface {
	HealthcheckUrl() string
	Bridge(clientConn net.Conn) error
	RemoveBridge(bridge Bridge) error
	RemoveAndCloseAllBridges()
	AddBridge(bridge Bridge)
	Dial() (net.Conn, error)
	Bridges() []Bridge
	IndexOfBridge(bridge Bridge) (int, error)
}

type backend struct {
	bridges         []Bridge
	Desc            string
	ipAddress       string
	port            uint
	healthcheckPort uint
	logger          lager.Logger
}

func NewBackend(desc, ipAddress string, port uint, healthcheckPort uint, logger lager.Logger) Backend {
	return &backend{
		Desc:            desc,
		bridges:         []Bridge{},
		ipAddress:       ipAddress,
		port:            port,
		healthcheckPort: healthcheckPort,
		logger:          logger,
	}
}

func (b *backend) HealthcheckUrl() string {
	endpoint := fmt.Sprintf("http://%s:%d", b.ipAddress, b.healthcheckPort)
	return endpoint
}

func (b *backend) Bridge(clientConn net.Conn) error {
	backendConn, err := b.Dial()
	if err != nil {
		return errors.New(fmt.Sprintf("Error connection to backend: %v", err))
	}

	bridge := NewConnectionBridge(clientConn, backendConn, b.logger)
	b.AddBridge(bridge)

	go func() {
		bridge.Connect()
		b.RemoveBridge(bridge)
	}()

	return nil
}

func (b *backend) RemoveBridge(bridge Bridge) error {
	index, err := b.IndexOfBridge(bridge)
	if err != nil {
		return err
	}
	b.removeBridgeAt(index)
	return nil
}

func (b *backend) RemoveAndCloseAllBridges() {
	for _, bridge := range b.bridges {
		bridge.Close()
	}
	b.bridges = []Bridge{}
}

func (b *backend) AddBridge(bridge Bridge) {
	b.bridges = append(b.bridges, bridge)
}

func (b *backend) Dial() (net.Conn, error) {
	addr := fmt.Sprintf("%s:%d", b.ipAddress, b.port)
	backendConn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return backendConn, nil
}

func (b *backend) Bridges() []Bridge {
	return b.bridges
}

func (b *backend) IndexOfBridge(bridge Bridge) (int, error) {
	index := -1
	for i, aBridge := range b.bridges {
		if aBridge == bridge {
			index = i
			break
		}
	}
	if index == -1 {
		return -1, errors.New("Bridge not found in backend")
	}
	return index, nil
}

func (b *backend) removeBridgeAt(index int) {
	copy(b.bridges[index:], b.bridges[index+1:])
	b.bridges = b.bridges[:len(b.bridges)-1]
}