package main

import (
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
)

func main() {
	cmdAddr := flag.String("cmd-address", "", "address to SSH into")
	psAddr := flag.String("ps-address", "", "address to SSH into")
	keyFile := flag.String("key-file", "", "location of SSH key")
	flag.Parse()
	key, err := ioutil.ReadFile(*keyFile)
	if err != nil {
		fmt.Printf("keyfile read error: %s\n", err)
		os.Exit(1)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		fmt.Printf("failed to parse private key: %s\n", err)
		os.Exit(1)
	}
	sshConfig := &ssh.ClientConfig{
		User: "Administrator",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	commands := []struct {
		name string
		cmd  string
	}{
		{
			name: "double quotes",
			cmd:  `powershell.exe -NonInteractive -ExecutionPolicy Bypass "if( Test-Path -Path / ){echo foo}"`,
		},
		{
			name: "escaped double quotes",
			cmd:  `powershell.exe -NonInteractive -ExecutionPolicy Bypass \"if( Test-Path -Path / ){echo foo}\"`,
		},
		{
			name: "single quotes",
			cmd:  `powershell.exe -NonInteractive -ExecutionPolicy Bypass 'if( Test-Path -Path / ){echo foo}'`,
		},
		{
			name: "escaped single quotes",
			cmd:  `powershell.exe -NonInteractive -ExecutionPolicy Bypass \'if( Test-Path -Path / ){echo foo}\'`,
		},
		{
			name: "-Command parameter with double quotes",
			cmd:  `powershell.exe -NonInteractive -ExecutionPolicy Bypass -Command "if( Test-Path -Path / ){echo foo}"`,
		},
		{
			name: "-Command parameter with escaped double quotes",
			cmd:  `powershell.exe -NonInteractive -ExecutionPolicy Bypass -Command \"if( Test-Path -Path / ){echo foo}\"`,
		},
		{
			name: "-Command parameter with single quotes",
			cmd:  `powershell.exe -NonInteractive -ExecutionPolicy Bypass -Command 'if( Test-Path -Path / ){echo foo}'`,
		},
		{
			name: "-Command parameter with escaped single quotes",
			cmd:  `powershell.exe -NonInteractive -ExecutionPolicy Bypass -Command \'if( Test-Path -Path / ){echo foo}\'`,
		},
	}
	for i, cmdCase := range commands {
		fmt.Printf("\n---- Case %d:\n", i)
		fmt.Printf("Testing %s\n", cmdCase.name)
		fmt.Printf("Running command: %s\n", cmdCase.cmd)
		fmt.Printf("Powershell VM results:\n")
		out, err := runCommandAgainst(*psAddr, cmdCase.cmd, sshConfig)
		fmt.Printf("Output: %s\n", out)
		fmt.Printf("Error: %s\n", err)
		fmt.Printf("CMD VM results:\n")
		out, err = runCommandAgainst(*cmdAddr, cmdCase.cmd, sshConfig)
		fmt.Printf("Output: %s\n", out)
		fmt.Printf("Error: %s\n", err)
	}
}

func runCommandAgainst(addr, cmd string, sshConfig *ssh.ClientConfig) (string, error) {
	sshClient, err := ssh.Dial("tcp", addr+":22", sshConfig)
	if err != nil {
		fmt.Printf("ssh dial error: %s\n", err)
		os.Exit(1)
	}
	sesh, err := sshClient.NewSession()
	if err != nil {
		fmt.Printf("session creation error: %s\n", err)
		os.Exit(1)
	}
	defer sesh.Close()
	out, err := sesh.CombinedOutput(cmd)
	if err != nil {
		return string(out), fmt.Errorf("command run error: %w\n", err)
	}
	return string(out), nil
}
