package multisite

type ServiceAPI interface {
	ID() string
	UpdateServiceAndWait(instanceName string, arbitraryParams string, planName *string) error
	CreateHostInfoKey(instanceName string) (key string, err error)
	CreateCredentialsKey(instanceName string) (key string, err error)
	InstanceExists(instanceName string) error
	InstancePlanName(instanceName string) (planName string, err error)
	PlanExists(planName string) (exists bool, err error)
}

type Logger interface {
	Printf(format string, v ...any)
}

type Workflow struct {
	Foundation1 ServiceAPI
	Foundation2 ServiceAPI
	Logger      Logger
}

func NewWorkflow(foundation1, foundation2 ServiceAPI, logger Logger) Workflow {
	return Workflow{
		Foundation1: foundation1,
		Foundation2: foundation2,
		Logger:      logger,
	}
}
