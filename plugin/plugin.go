package plugin

import (
	"code.cloudfoundry.org/cli/plugin"
	"io"
	"log"
	"time"
)

type Mysql struct {
	connectionWrapper ConnectionWrapper
	exiter            Exiter
	stdout            io.Writer
	logWaitTimeout    time.Duration
}

func New(connectionWrapper ConnectionWrapper, exiter Exiter, stdout io.Writer, logWaitTimeout time.Duration) *Mysql {
	return &Mysql{
		connectionWrapper: connectionWrapper,
		exiter:            exiter,
		stdout:            stdout,
		logWaitTimeout:    logWaitTimeout,
	}
}

//go:generate counterfeiter . Exiter
type Exiter interface {
	Exit(code int)
}

//go:generate counterfeiter . ConnectionWrapper
type ConnectionWrapper interface {
	IsSpaceDeveloper(plugin.CliConnection) (bool, error)
	UnpackAssets() (assetDir string, err error)
	PushApp(connection plugin.CliConnection, assetDir string) (appName string, err error)
	ExecuteMigrateTask(connection plugin.CliConnection, appName, sourceServiceName, destinationServiceName  string) (state string, err error)
	ShowRecentLogs(connection plugin.CliConnection, writer io.Writer) error
	Cleanup(connection plugin.CliConnection) error
}

func (p *Mysql) Run(cliConnection plugin.CliConnection, args []string) {
	var (
		logger            = log.New(p.stdout, "[MIGRATE-SPIKE] ", log.LstdFlags)
		sourceServiceName = args[2]
		destServiceName   = args[3]
	)

	defer p.connectionWrapper.Cleanup(cliConnection)

	ok, err := p.connectionWrapper.IsSpaceDeveloper(cliConnection)
	if err != nil {
		logger.Println(err)
		p.exiter.Exit(1)
		return
	}

	if !ok {
		logger.Println("You must have the 'Space Developer' privilege to use the 'cf mysql migrate' command")
		p.exiter.Exit(1)
		return
	}

	assetsDir, err := p.connectionWrapper.UnpackAssets()
	if err != nil {
		logger.Println(err)
		p.exiter.Exit(1)
		return
	}

	appName, err := p.connectionWrapper.PushApp(cliConnection, assetsDir)
	if err != nil {
		logger.Println(err)
		p.exiter.Exit(1)
		return
	}

	state, err := p.connectionWrapper.ExecuteMigrateTask(cliConnection, appName, sourceServiceName, destServiceName)
	if err != nil {
		logger.Println(err)
		p.exiter.Exit(1)
		return
	}

	if state != "SUCCEEDED" {
		logger.Println("Migration failed. Fetching log output...")
		time.Sleep(p.logWaitTimeout)
		p.connectionWrapper.ShowRecentLogs(cliConnection, p.stdout)
		p.exiter.Exit(1)
		return
	}

	p.exiter.Exit(0)
}
