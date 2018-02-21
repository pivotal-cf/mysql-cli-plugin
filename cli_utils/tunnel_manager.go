package cli_utils

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
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

func NewTunnelManager(cfCommandRunner CfCommandRunner, tunnels []Tunnel) *TunnelManager {
	for _, tunnel := range tunnels {

		connectionString := fmt.Sprintf(
			"%s:%s@tcp(127.0.0.1:%d)/%s?interpolateParams=true&tls=skip-verify",
			tunnel.ServiceKey.Username,
			tunnel.ServiceKey.Password,
			tunnel.Port,
			tunnel.ServiceKey.DBName,
		)
		db, err := sql.Open("mysql", connectionString)
		if err != nil {
			log.Fatalf("Error creating database connection: %v", err)
		}

		tunnel.DB = db
	}

	return &TunnelManager{
		CmdRunner: cfCommandRunner,
		Tunnels: tunnels,
	}
}

func (t *TunnelManager) CreateSSHTunnel() error {
	// TODO assign ephemeral ports instead of having them sent in from the caller
	args := []string{"ssh", t.AppName, "-N"}

	for _, spec := range t.Tunnels {
		tunnelSpec := fmt.Sprintf("%d:%s:3306", spec.Port, spec.ServiceKey.Hostname)
		args = append(args, "-L", tunnelSpec)
	}
	log.Printf("Opening tunnel with `%s`", strings.Join(args, " "))
	_, err := t.CmdRunner.CliCommandWithoutTerminalOutput(args...)
	if err != nil {
		return err
	}

	err = t.WaitForTunnel(60 * time.Second)
	if err != nil {
		return err
	}

	return nil
}

func (t *TunnelManager) WaitForTunnel(timeout time.Duration) error {

	timerCh := time.After(timeout)
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-timerCh:
			return errors.New("Timeout")
		case <-ticker.C:
			var unused int
			var tunnels []Tunnel
			copy(tunnels, t.Tunnels)
			for i, tunnel := range tunnels {
				if err := tunnel.DB.QueryRow("SELECT 1").Scan(&unused); err == nil {
					tunnels = append(tunnels[:i], tunnels[i+1:]...)
				}
			}

			if len(tunnels) == 0 {
				return nil
			}
		}
	}

	return nil
}
