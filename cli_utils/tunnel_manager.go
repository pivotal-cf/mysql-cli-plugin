package cli_utils

import (
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"
)

type Tunnel struct {
	DB         *sql.DB
	Port       int
	ServiceKey ServiceKey
}

type TunnelManager struct {
	Tunnels   []Tunnel
	CmdRunner CfCommandRunner
	AppName   string
}

func NewTunnelManager(cfCommandRunner CfCommandRunner, tunnels []Tunnel) (*TunnelManager, error) {
	for idx, tunnel := range tunnels {
		addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, err
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return nil, err
		}
		port := l.Addr().(*net.TCPAddr).Port
		l.Close()

		connectionString := fmt.Sprintf(
			"%s:%s@tcp(127.0.0.1:%d)/%s?interpolateParams=true&tls=skip-verify",
			tunnel.ServiceKey.Username,
			tunnel.ServiceKey.Password,
			port,
			tunnel.ServiceKey.DBName,
		)
		db, err := sql.Open("mysql", connectionString)
		if err != nil {
			return nil, errors.Wrapf(err, "Error creating database connection for connection %s", connectionString)
		}

		tunnels[idx].DB = db
		tunnels[idx].Port = port
	}
	return &TunnelManager{
		CmdRunner: cfCommandRunner,
		Tunnels:   tunnels,
	}, nil
}

func (t *TunnelManager) CreateSSHTunnel() error {
	args := []string{"ssh", t.AppName, "-N"}

	for _, spec := range t.Tunnels {
		tunnelSpec := fmt.Sprintf("%d:%s:3306", spec.Port, spec.ServiceKey.Hostname)
		args = append(args, "-L", tunnelSpec)
	}
	_, err := t.CmdRunner.CliCommandWithoutTerminalOutput(args...)
	if err != nil && err.Error() != "Error: EOF" {
		return errors.Wrapf(err, "Failed to open ssh tunnel to app %s", t.AppName)
	}
	return nil
}

func (t *TunnelManager) WaitForTunnel(timeout time.Duration) error {
	timerCh := time.After(timeout)
	ticker := time.NewTicker(1 * time.Second)
	tunnelStatuses := make([]bool, len(t.Tunnels))

	var allTunnelsOpen = func() bool {
		for _, status := range tunnelStatuses {
			if !status {
				return false
			}
		}

		return true
	}

	for {
		select {
		case <-timerCh:
			return errors.New("Timeout")
		case <-ticker.C:
			for i, tunnel := range t.Tunnels {
				var unused int
				if err := tunnel.DB.QueryRow("SELECT 1").Scan(&unused); err == nil {
					tunnelStatuses[i] = true
				}
			}

			if allTunnelsOpen() {
				return nil
			}
		}
	}

	return nil
}
