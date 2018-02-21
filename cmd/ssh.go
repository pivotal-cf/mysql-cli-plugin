package main

// Example on how to set up SSH tunnels to 2 different remote endpoints

import (
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	sshHost    = "ssh.bosh-lite.com:2222"
	appGuid    = "1ff215c1-5389-4005-87b9-30075413ab63"
	userPrefix = "cf"
	instance   = "0"
	localAddr  = "127.0.0.1:63306"
	localAddr2  = "127.0.0.1:63307"

	remoteAddr = "10.244.16.5:3306"
	remoteAddr2 = "10.244.16.4:3306"
)

func getCode() string {
	out, err := exec.Command("cf", "ssh-code").CombinedOutput()
	if err != nil {
		log.Fatalf("Error getting code: %v", err)
	}

	return string(out)
}

func handleForwardConnection(secureClient *ssh.Client, conn net.Conn, targetAddr string) {
	defer conn.Close()

	target, err := secureClient.Dial("tcp", targetAddr)
	if err != nil {
		fmt.Printf("connect to %s failed: %s\n", targetAddr, err.Error())
		return
	}
	defer target.Close()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go copyAndClose(wg, conn, target)
	go copyAndClose(wg, target, conn)
	wg.Wait()
}

func localForwardAcceptLoop(secureClient *ssh.Client, listener net.Listener, addr string) {
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return
		}

		go handleForwardConnection(secureClient, conn, addr)
	}
}

func copyAndClose(wg *sync.WaitGroup, dest io.WriteCloser, src io.Reader) {
	_, _ = io.Copy(dest, src)
	_ = dest.Close()
	if wg != nil {
		wg.Done()
	}
}

func main() {

	passcode := strings.TrimSpace(getCode())

	user := fmt.Sprintf("%s:%s/%s", userPrefix, appGuid, instance)

	log.Println(user)
	log.Println(passcode)

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(passcode),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	secureClient, err := ssh.Dial("tcp", sshHost, sshConfig)
	if err != nil {
		log.Fatalf("Error connecting to %s@%s: %v", user, sshHost, err)
	}
	defer secureClient.Close()

	local, err := net.Listen("tcp", localAddr)
	if err != nil {
		log.Fatalf("Error listening on local %s: %v", localAddr, err)
	}
	defer local.Close()

	local2, err := net.Listen("tcp", localAddr2)
	if err != nil {
		log.Fatalf("Error listening on local %s: %v", localAddr2, err)
	}
	defer local2.Close()


	var wg sync.WaitGroup
	wg.Add(1)
	go localForwardAcceptLoop(secureClient, local, remoteAddr)
	go localForwardAcceptLoop(secureClient, local2, remoteAddr2)
	wg.Wait()
}
