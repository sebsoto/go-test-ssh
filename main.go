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
	setPSShell := flag.Bool("set-ps", false, "if true, the VM specified by ps-address will be configured"+
		"to have powershell as the default shell, before any other commands are ran")
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
			name: "-Command parameter with script block",
			cmd:  `powershell.exe -NonInteractive -ExecutionPolicy Bypass -Command {if( Test-Path -Path / ){echo foo}}`,
		},
		{
			name: "-Command parameter with quoted script block",
			cmd:  `powershell.exe -NonInteractive -ExecutionPolicy Bypass -Command "{if( Test-Path -Path / ){echo foo}}"`,
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
		//{
		//	name: "Running CMD commands, with double quotes - cmd passes",
		//	cmd:  `sc.exe create hybrid-overlay-node binPath="C:\\k\\hybrid-overlay-node.exe --node winhost --hybrid-overlay-vxlan-port=9898 --k8s-kubeconfig c:\\k\\kubeconfig --windows-service --logfile C:\\var\\log\\hybrid-overlay\\hybrid-overlay.log" start=auto depend=kubelet`,
		//},
		//{
		//	name: "Running CMD commands, with single quotes - ps passes",
		//	cmd:  `sc.exe create hybrid-overlay-node binPath='C:\\k\\hybrid-overlay-node.exe --node winhost --hybrid-overlay-vxlan-port=9898 --k8s-kubeconfig c:\\k\\kubeconfig --windows-service --logfile C:\\var\\log\\hybrid-overlay\\hybrid-overlay.log' start=auto depend=kubelet`,
		//},
		//{
		//	name: "Running CMD commands with cmd/c, with double quotes - cmd passes",
		//	cmd:  `cmd /c sc.exe create hybrid-overlay-node binPath="C:\\k\\hybrid-overlay-node.exe --node winhost --hybrid-overlay-vxlan-port=9898 --k8s-kubeconfig c:\\k\\kubeconfig --windows-service --logfile C:\\var\\log\\hybrid-overlay\\hybrid-overlay.log" start=auto depend=kubelet`,
		//},
		//{
		//	name: "Running CMD commands with cmd /c, with single quotes - ps passes",
		//	cmd:  `cmd /c sc.exe create hybrid-overlay-node binPath='C:\\k\\hybrid-overlay-node.exe --node winhost --hybrid-overlay-vxlan-port=9898 --k8s-kubeconfig c:\\k\\kubeconfig --windows-service --logfile C:\\var\\log\\hybrid-overlay\\hybrid-overlay.log' start=auto depend=kubelet`,
		//},
	}
	if *setPSShell {
		setPSCMD := `powershell.exe -NonInteractive -ExecutionPolicy Bypass New-ItemProperty -Path "HKLM:\SOFTWARE\OpenSSH" -Name DefaultShell -Value "C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe" -PropertyType String -Force`
		out, err := runCommandAgainst(*psAddr, setPSCMD, sshConfig)
		if err != nil {
			fmt.Printf("error setting PS default shell: %s %s", out, err)
			os.Exit(1)
		}
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
