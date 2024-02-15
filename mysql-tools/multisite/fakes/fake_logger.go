package fakes

import (
	"fmt"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite"
)

type FakeLogger struct {
	Operations *[]string
}

func (l *FakeLogger) Printf(format string, v ...any) {
	op := fmt.Sprintf("logger.Printf(%q)", fmt.Sprintf(format, v...))
	*l.Operations = append(*l.Operations, op)
}

var _ multisite.Logger = (*FakeLogger)(nil)
