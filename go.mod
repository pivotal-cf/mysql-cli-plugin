module github.com/pivotal-cf/mysql-cli-plugin

go 1.21.1

require (
	code.cloudfoundry.org/cli v7.1.0+incompatible
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/blang/semver/v4 v4.0.0
	github.com/cloudfoundry-community/go-cfclient/v2 v2.0.0
	github.com/cloudfoundry/cf-test-helpers/v2 v2.9.0
	github.com/go-sql-driver/mysql v1.8.1
	github.com/google/uuid v1.6.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/jessevdk/go-flags v1.6.1
	github.com/maxbrunsfeld/counterfeiter/v6 v6.8.1
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/ginkgo/v2 v2.19.0
	github.com/onsi/gomega v1.33.1
	github.com/pivotal-cf/go-binmock v0.0.0-20171027112700-f797157c64e9
)

require (
	code.cloudfoundry.org/bytefmt v0.0.0-20231017140541-3b893ed0421b // indirect
	code.cloudfoundry.org/cli-plugin-repo v0.0.0-20240118200324-f1484a8aa754 // indirect
	code.cloudfoundry.org/go-log-cache v1.0.0 // indirect
	code.cloudfoundry.org/go-loggregator v7.4.0+incompatible // indirect
	code.cloudfoundry.org/gofileutils v0.0.0-20170111115228-4d0c80011a0f // indirect
	code.cloudfoundry.org/jsonry v1.1.4 // indirect
	code.cloudfoundry.org/rfc5424 v0.0.0-20201103192249-000122071b78 // indirect
	code.cloudfoundry.org/tlsconfig v0.0.0-20240213144909-765c8d6ec2ff // indirect
	code.cloudfoundry.org/ykk v0.0.0-20170424192843-e4df4ce2fd4d // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/SermoDigital/jose v0.9.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/bmizerany/pat v0.0.0-20210406213842-e4b6760bdd6f // indirect
	github.com/charlievieth/fs v0.0.3 // indirect
	github.com/cloudfoundry/bosh-cli v6.4.1+incompatible // indirect
	github.com/cloudfoundry/bosh-utils v0.0.442 // indirect
	github.com/cppforlife/go-patch v0.2.0 // indirect
	github.com/cyphar/filepath-securejoin v0.2.4 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/pprof v0.0.0-20240424215950-a892ee059fd6 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/kr/pty v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/moby/moby v0.0.0-00010101000000-000000000000 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/sabhiram/go-gitignore v0.0.0-20210923224102-525f6e181f06 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/tedsuo/rata v1.0.0 // indirect
	github.com/vito/go-interact v1.0.0 // indirect
	golang.org/x/crypto v0.23.0 // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/oauth2 v0.14.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/term v0.20.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	golang.org/x/tools v0.21.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240213162025-012b6fc9bca9 // indirect
	google.golang.org/grpc v1.61.1 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/cheggaaa/pb.v1 v1.0.28 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

exclude github.com/vito/go-interact v1.0.1

replace github.com/moby/moby => github.com/moby/moby v20.10.25+incompatible
