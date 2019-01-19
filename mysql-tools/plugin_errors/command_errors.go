package plugin_errors

import "fmt"

type UsageError struct {
	Msg, Usage string
}

func (ue UsageError) Error() string {
	return fmt.Sprintf("%s\n%s", ue.Msg, ue.Usage)
}

func NewUsageError(text string) *UsageError {
	usage := `NAME:
   mysql-tools - Plugin to migrate mysql instances

USAGE:
   cf mysql-tools migrate [-h] [--no-cleanup] <source-service-instance> <p.mysql-plan-type>
   cf mysql-tools version`
	return &UsageError{Usage: usage, Msg: text}
}

type CustomUsageError struct {
	Msg, Usage string
}

func (ue CustomUsageError) Error() string {
	return fmt.Sprintf("%s\n%s", ue.Msg, ue.Usage)
}

func NewCustomUsageError(text, usage string) *CustomUsageError {
	return &CustomUsageError{Usage: usage, Msg: text}
}
