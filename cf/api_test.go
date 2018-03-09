package cf_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/mysql-cli-plugin/cf"
	"github.com/pivotal-cf/mysql-cli-plugin/cf/cffakes"
)

var _ = Describe("GetAppByName", func() {
	var (
		client              *cf.Api
		fakeCfCommandRunner *cffakes.FakeCfCommandRunner
	)

	BeforeEach(func() {
		fakeCfCommandRunner = new(cffakes.FakeCfCommandRunner)
		client = cf.NewApi(fakeCfCommandRunner)
	})

	It("returns an application by its name", func() {
		fakeCfCommandRunner.CliCommandWithoutTerminalOutputReturns([]string{
			`{"resources": [{`,
			`"guid": "6aef0cf0-c5d5-4ec1-89ae-73971d24241c",`,
			`"name": "mysql-migrate"`,
			`}]}`,
		}, nil)

		app, err := client.GetAppByName("mysql-migrate")
		Expect(err).NotTo(HaveOccurred())
		Expect(app.Name).To(Equal("mysql-migrate"))
		Expect(app.Guid).To(Equal("6aef0cf0-c5d5-4ec1-89ae-73971d24241c"))

		args := fakeCfCommandRunner.CliCommandWithoutTerminalOutputArgsForCall(0)
		Expect(args).To(Equal([]string{"curl", "/v3/apps?names=mysql-migrate"}))
	})

	Context("when there is an error getting an application by name", func() {
		It("returns an error", func() {
			fakeCfCommandRunner.CliCommandWithoutTerminalOutputReturns(
				nil, errors.New("some-error"))

			_, err := client.GetAppByName("mysql-migrate")
			Expect(err).To(MatchError("failed to retrieve an app by name: some-error"))
		})
	})

	Context("when there are no applications returned by name", func() {
		It("returns an error", func() {
			fakeCfCommandRunner.CliCommandWithoutTerminalOutputReturns([]string{
				`{"resources": [`,
				`]}`,
			}, nil)

			_, err := client.GetAppByName("mysql-migrate")
			Expect(err).To(MatchError("failed to retrieve an app by name: none found"))
		})
	})

	Context("when invalid json is returned", func() {
		It("returns an error with the contents", func() {
			fakeCfCommandRunner.CliCommandWithoutTerminalOutputReturns([]string{
				`something bad happened`,
			}, nil)

			_, err := client.GetAppByName("mysql-migrate")
			Expect(err).To(MatchError("failed to parse the following api response: something bad happened"))
		})
	})
})

var _ = Describe("GetTaskByGUID", func() {
	var (
		client              *cf.Api
		fakeCfCommandRunner *cffakes.FakeCfCommandRunner
	)

	BeforeEach(func() {
		fakeCfCommandRunner = new(cffakes.FakeCfCommandRunner)
		client = cf.NewApi(fakeCfCommandRunner)
	})

	It("returns an task by its guid", func() {
		fakeCfCommandRunner.CliCommandWithoutTerminalOutputReturns([]string{
			`{`,
			`"guid": "6aef0cf0-c5d5-4ec1-89ae-73971d24241c",`,
			`"state": "RUNNING"`,
			`}`,
		}, nil)

		task, err := client.GetTaskByGUID("6aef0cf0-c5d5-4ec1-89ae-73971d24241c")
		Expect(err).NotTo(HaveOccurred())
		Expect(task.State).To(Equal("RUNNING"))
		Expect(task.Guid).To(Equal("6aef0cf0-c5d5-4ec1-89ae-73971d24241c"))

		args := fakeCfCommandRunner.CliCommandWithoutTerminalOutputArgsForCall(0)
		Expect(args).To(Equal([]string{"curl", "/v3/tasks/6aef0cf0-c5d5-4ec1-89ae-73971d24241c"}))
	})

	Context("when there is an error getting a task by guid", func() {
		It("returns an error", func() {
			fakeCfCommandRunner.CliCommandWithoutTerminalOutputReturns(
				nil, errors.New("some-error"))

			_, err := client.GetTaskByGUID("6aef0cf0-c5d5-4ec1-89ae-73971d24241c")
			Expect(err).To(MatchError("failed to retrieve a task by guid: some-error"))
		})
	})

	Context("when invalid json is returned", func() {
		It("returns an error with the contents", func() {
			fakeCfCommandRunner.CliCommandWithoutTerminalOutputReturns([]string{
				`something bad happened`,
			}, nil)

			_, err := client.GetTaskByGUID("6aef0cf0-c5d5-4ec1-89ae-73971d24241c")
			Expect(err).To(MatchError("failed to parse the following api response: something bad happened"))
		})
	})
})

var _ = Describe("CreateTask", func() {
	var (
		client              *cf.Api
		fakeCfCommandRunner *cffakes.FakeCfCommandRunner
	)

	BeforeEach(func() {
		fakeCfCommandRunner = new(cffakes.FakeCfCommandRunner)
		client = cf.NewApi(fakeCfCommandRunner)
	})

	It("creates a task", func() {
		fakeCfCommandRunner.CliCommandWithoutTerminalOutputReturns([]string{
			`{`,
			`"guid": "abc-123",`,
			`"state": "RUNNING"`,
			`}`,
		}, nil)

		task, err := client.CreateTask(cf.App{
			Guid: "6aef0cf0-c5d5-4ec1-89ae-73971d24241c",
		}, "some-command")
		Expect(err).NotTo(HaveOccurred())
		Expect(task.State).To(Equal("RUNNING"))
		Expect(task.Guid).To(Equal("abc-123"))

		args := fakeCfCommandRunner.CliCommandWithoutTerminalOutputArgsForCall(0)
		Expect(args).To(Equal([]string{
			"curl", "-X", "POST", "-d",
			`{"command":"some-command"}`,
			"/v3/apps/6aef0cf0-c5d5-4ec1-89ae-73971d24241c/tasks"}))
	})

	Context("when there is an error creating the task", func() {
		It("returns an error", func() {
			fakeCfCommandRunner.CliCommandWithoutTerminalOutputReturns(
				nil, errors.New("some-error"))

			_, err := client.CreateTask(cf.App{
				Guid: "6aef0cf0-c5d5-4ec1-89ae-73971d24241c",
			}, "some-command")
			Expect(err).To(MatchError("failed to create a task: some-error"))
		})
	})

	Context("when invalid json is returned", func() {
		It("returns an error with the contents", func() {
			fakeCfCommandRunner.CliCommandWithoutTerminalOutputReturns([]string{
				`something bad happened`,
			}, nil)

			_, err := client.CreateTask(cf.App{
				Guid: "6aef0cf0-c5d5-4ec1-89ae-73971d24241c",
			}, "some-command")
			Expect(err).To(MatchError("failed to parse the following api response: something bad happened"))
		})
	})

	Context("when the returned task has errors", func() {
		It("returns the errors", func() {
			fakeCfCommandRunner.CliCommandWithoutTerminalOutputReturns([]string{
				`{`,
				`"guid": "abc-123",`,
				`"state": "RUNNING",`,
				`"errors": [{`,
				`"detail": "some-detail",`,
				`"title": "some-title",`,
				`"code": 404`,
				`}]`,
				`}`,
			}, nil)

			_, err := client.CreateTask(cf.App{
				Guid: "6aef0cf0-c5d5-4ec1-89ae-73971d24241c",
			}, "some-command")
			Expect(err).To(MatchError("failed to create a task: 404: some-title - some-detail"))
		})

	})
})

var _ = Describe("WaitForTask", func() {
	var (
		client              *cf.Api
		fakeCfCommandRunner *cffakes.FakeCfCommandRunner
	)

	BeforeEach(func() {
		fakeCfCommandRunner = new(cffakes.FakeCfCommandRunner)
		client = cf.NewApi(fakeCfCommandRunner)
	})

	Context("when the task succeeds", func() {
		It("returns no error", func() {
			task := cf.Task{Guid: "some-guid"}
			fakeCfCommandRunner.CliCommandWithoutTerminalOutputReturns([]string{
				`{`,
				`"guid": "some-guid",`,
				`"state": "SUCCEEDED"`,
				`}`,
			}, nil)

			Expect(client.WaitForTask(task)).To(Equal("SUCCEEDED"))
		})
	})

	Context("when the task fails", func() {
		It("returns the error", func() {
			task := cf.Task{Guid: "some-guid"}
			fakeCfCommandRunner.CliCommandWithoutTerminalOutputReturns(nil, errors.New("some-error"))

			_, err := client.WaitForTask(task)
			Expect(err).To(MatchError("failed to retrieve a task by guid: some-error"))
		})
	})

	Context("when polling the task has a network blip and eventually succeeds", func() {
		It("returns no error", func() {

			fakeCfCommandRunner.CliCommandWithoutTerminalOutputReturnsOnCall(0,
				[]string{`{"guid": "some-guid", "state": "RUNNING"}`}, nil,
			)
			fakeCfCommandRunner.CliCommandWithoutTerminalOutputReturnsOnCall(1,
				[]string{`{"errors":[{"detail":"Invalid Auth Token","title":"CF-InvalidAuthToken","code":1000}]}`}, nil,
			)
			fakeCfCommandRunner.CliCommandWithoutTerminalOutputReturnsOnCall(2,
				[]string{`{"guid": "some-guid", "state": "SUCCEEDED"}`}, nil,
			)

			task := cf.Task{Guid: "some-guid"}
			Expect(client.WaitForTask(task)).To(Equal("SUCCEEDED"))

			Expect(fakeCfCommandRunner.CliCommandWithoutTerminalOutputArgsForCall(0)).To(Equal(
				[]string{"curl", "/v3/tasks/some-guid"},
			))
			Expect(fakeCfCommandRunner.CliCommandWithoutTerminalOutputArgsForCall(1)).To(Equal(
				[]string{"curl", "/v3/tasks/some-guid"},
			))
			Expect(fakeCfCommandRunner.CliCommandWithoutTerminalOutputArgsForCall(2)).To(Equal(
				[]string{"curl", "/v3/tasks/some-guid"},
			))
		})
	})
})
