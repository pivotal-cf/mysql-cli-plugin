package ssh

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/pivotal-cf/mysql-v2-cli-plugin/service"
)

//go:generate counterfeiter . CfCommandRunner
type CfCommandRunner interface {
	CliCommand(...string) ([]string, error)
	CliCommandWithoutTerminalOutput(args ...string) ([]string, error)
}

//go:generate counterfeiter . DB
type DB interface {
	Ping(serviceInfo *service.ServiceInfo, port int) error
}

var (
	// TODO don't hardcode app name
	appName    = "static-app"
	staticFile = "Staticfile"
	indexFile  = "index.html"
)

type TunnelManager struct {
	cfCommandRunner CfCommandRunner
	db              DB
	tmpDir          string
	dbPollTimeout   time.Duration
}

func NewTunnerManager(cfCommandRunner CfCommandRunner, db DB, tmpDir string, dbPollTimeout time.Duration) *TunnelManager {
	return &TunnelManager{
		cfCommandRunner: cfCommandRunner,
		db:              db,
		tmpDir:          tmpDir,
		dbPollTimeout:   dbPollTimeout,
	}
}

func (m *TunnelManager) Start(servicesInfo ...*service.ServiceInfo) error {
	err := m.pushApp()
	if err != nil {
		return err
	}

	ports, err := m.getFreePorts(len(servicesInfo))
	if err != nil {
		return err
	}

	go m.createSSHTunnel(servicesInfo, ports)

	err = m.waitForSSHTunnel(servicesInfo, ports)
	if err != nil {
		return err
	}

	for i, _ := range servicesInfo {
		servicesInfo[i].LocalSSHPort = ports[i]
	}

	return nil
}

func (m *TunnelManager) Close() {
	m.cfCommandRunner.CliCommand("delete", appName, "-f")
}

func (m *TunnelManager) waitForSSHTunnel(servicesInfo []*service.ServiceInfo, ports []int) error {
	var (
		timerCh        = time.After(m.dbPollTimeout)
		ticker         = time.NewTicker(time.Second)
		tunnelStatuses = make([]bool, len(servicesInfo))
		allTunnelsOpen = func() bool {
			for _, status := range tunnelStatuses {
				if !status {
					return false
				}
			}

			return true
		}
	)

	for {
		select {
		case <-timerCh:
			return errors.New("timeout")
		case <-ticker.C:
			for i, serviceInfo := range servicesInfo {
				err := m.db.Ping(serviceInfo, ports[i])
				if err == nil {
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

func (m *TunnelManager) createSSHTunnel(servicesInfo []*service.ServiceInfo, ports []int) {
	args := []string{"ssh", appName, "-N"}

	for i, serviceInfo := range servicesInfo {
		tunnelSpec := fmt.Sprintf("%d:%s:3306", ports[i], serviceInfo.Hostname)
		args = append(args, "-L", tunnelSpec)
	}

	m.cfCommandRunner.CliCommandWithoutTerminalOutput(args...)
}

func (m *TunnelManager) pushApp() error {
	appDir := filepath.Join(m.tmpDir, appName)

	err := os.Mkdir(appDir, 0700)
	if err != nil {
		return fmt.Errorf("failed to create app directory: %s", err)
	}

	err = ioutil.WriteFile(filepath.Join(appDir, staticFile), nil, 0600)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(appDir, indexFile), nil, 0600)
	if err != nil {
		return err
	}

	_, err = m.cfCommandRunner.CliCommand("push", appName, "--random-route", "-b", "staticfile_buildpack", "-p", appDir)
	if err != nil {
		return fmt.Errorf("failed to push application: %s", err)
	}

	return nil
}

func (m *TunnelManager) getFreePorts(count int) ([]int, error) {
	var ports []int

	for i := 0; i < count; i++ {
		port, err := m.getFreePort()
		if err != nil {
			return nil, fmt.Errorf("failed to get free port: %s", err)
		}

		ports = append(ports, port)
	}

	return ports, nil
}

func (m *TunnelManager) getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}

	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
