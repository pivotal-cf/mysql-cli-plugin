module github.com/pivotal-cf/mysql-cli-plugin

go 1.21.1

require (
	code.cloudfoundry.org/cli v7.1.0+incompatible
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/blang/semver/v4 v4.0.0
	github.com/cloudfoundry-community/go-cfclient/v2 v2.0.0
	github.com/cloudfoundry/cf-test-helpers/v2 v2.7.0
	github.com/go-sql-driver/mysql v1.7.1
	github.com/google/uuid v1.3.1
	github.com/hashicorp/go-multierror v1.1.1
	github.com/jessevdk/go-flags v1.5.0
	github.com/maxbrunsfeld/counterfeiter/v6 v6.7.0
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/ginkgo/v2 v2.12.1
	github.com/onsi/gomega v1.27.10
	github.com/pivotal-cf/go-binmock v0.0.0-20171027112700-f797157c64e9
)

require (
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/pprof v0.0.0-20230912144702-c363fe2c2ed8 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/net v0.15.0 // indirect
	golang.org/x/oauth2 v0.10.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/tools v0.13.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

exclude github.com/vito/go-interact v1.0.1

replace github.com/moby/moby => github.com/moby/moby v20.10.25+incompatible
