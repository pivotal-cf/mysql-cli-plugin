package docker

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strings"

	"github.com/onsi/ginkgo"
)

func Command(args ...string) (string, error) {
	var out bytes.Buffer
	cmd := exec.Command("docker", args...)
	cmd.Stdout = io.MultiWriter(ginkgo.GinkgoWriter, &out)
	cmd.Stderr = io.MultiWriter(ginkgo.GinkgoWriter, &out)

	_, _ = fmt.Fprintln(ginkgo.GinkgoWriter, "$", strings.Join(cmd.Args, " "))
	err := cmd.Run()
	trimmedOutput := strings.TrimSpace(out.String())
	return trimmedOutput, err
}

func Run(args ...string) (string, error) {
	return Command(append([]string{"run", "--platform=linux/amd64"}, args...)...)
}

func CreateNetwork(name string) error {
	_, err := Command("network", "create", name)
	return err
}

func RemoveNetwork(name string) error {
	_, err := Command("network", "remove", name)
	return err
}

type ContainerSpec struct {
	Name        string
	Image       string
	Network     string
	Env         []string
	Volumes     []string
	Args        []string
	Entrypoint  string
	ExposePorts []string
}

func CreateContainer(spec ContainerSpec) (string, error) {
	args := []string{
		"--detach",
		"--publish-all",
		"--platform=linux/amd64",
	}

	if spec.Name != "" {
		args = append(args, "--name="+spec.Name)
	}

	if spec.Entrypoint != "" {
		args = append(args, "--entrypoint="+spec.Entrypoint)
	}

	if spec.Network != "" {
		args = append(args, "--network="+spec.Network)
	}

	for _, e := range spec.Env {
		args = append(args, "--env="+e)
	}

	for _, v := range spec.Volumes {
		args = append(args, "--volume="+v)
	}

	for _, p := range spec.ExposePorts {
		args = append(args, "--expose="+p)
	}

	args = append(args, spec.Image)
	args = append(args, spec.Args...)

	return Run(args...)
}

func RemoveContainer(name string) error {
	_, err := Command("container", "rm", "--force", "--volumes", name)
	return err
}

func ContainerPort(containerID, portSpec string) (string, error) {
	hostPort, err := Command("container", "port", containerID, portSpec)
	if err != nil {
		return "", err
	}

	_, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return "", err
	}

	return port, nil
}

func Logs(containerID string) error {
	_, _ = fmt.Fprintln(ginkgo.GinkgoWriter, "$ docker logs ", containerID)
	cmd := exec.Command("docker", "logs", containerID)
	cmd.Stdout = ginkgo.GinkgoWriter
	cmd.Stderr = ginkgo.GinkgoWriter
	return cmd.Run()
}
