package ssh

import (
	"bytes"
	"crypto"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Keys struct {
	Public  string
	Private string
}

func GenerateKey() (*Keys, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, err
	}
	p, err := ssh.MarshalPrivateKey(crypto.PrivateKey(priv), "")
	if err != nil {
		return nil, err
	}
	privateKeyPem := pem.EncodeToMemory(p)
	privateKeyString := string(privateKeyPem)
	publicKey, err := ssh.NewPublicKey(pub)
	if err != nil {
		return nil, err
	}
	publicKeyString := "ssh-ed25519" + " " + base64.StdEncoding.EncodeToString(publicKey.Marshal())
	return &Keys{
		Private: privateKeyString,
		Public:  publicKeyString,
	}, nil
}

func RunAgent(sshPrivateKey string) error {
	sshAgent := exec.Command("ssh-agent", "-s")
	var out bytes.Buffer
	sshAgent.Stdout = &out
	err := sshAgent.Run()

	if err != nil {
		log.Println(err)
		return err
	}
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		if strings.Contains(line, "SSH_AUTH_SOCK") {
			parts := strings.Split(line, ";")
			if len(parts) > 0 {
				keyAndValue := strings.Split(parts[0], "=")
				if len(keyAndValue) > 1 {
					os.Setenv(keyAndValue[0], keyAndValue[1])
				}
			}
		}
		if strings.Contains(line, "SSH_AGENT_PID") {
			parts := strings.Split(line, ";")
			if len(parts) > 0 {
				keyAndValue := strings.Split(parts[0], "=")
				if len(keyAndValue) > 1 {
					os.Setenv(keyAndValue[0], keyAndValue[1])
				}
			}
		}
	}

	sshAdd := exec.Command("ssh-add", sshPrivateKey)
	var out1 bytes.Buffer
	var stderr1 bytes.Buffer
	sshAdd.Stdout = &out1
	sshAdd.Stderr = &stderr1
	err = sshAdd.Run()
	if err != nil {
		fmt.Println("Executing: " + sshAdd.String())
		fmt.Println(fmt.Sprint(err) + ": " + stderr1.String())
		return err
	}
	return nil
}

/*
	sshKey := "/home/javier/.ssh/id_ed25519"
	sshAgent := exec.Command("ssh-agent", "-s")
	var out bytes.Buffer
	sshAgent.Stdout = &out
	err := sshAgent.Run()

	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("==== ssh agent executed ===")
	fmt.Println(out.String())

	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		if strings.Contains(line, "SSH_AUTH_SOCK") {
			parts := strings.Split(line, ";")
			if len(parts) > 0 {
				keyAndValue := strings.Split(parts[0], "=")
				if len(keyAndValue) > 1 {
					os.Setenv(keyAndValue[0], keyAndValue[1])
				}
			}
		}
		if strings.Contains(line, "SSH_AGENT_PID") {
			parts := strings.Split(line, ";")
			if len(parts) > 0 {
				keyAndValue := strings.Split(parts[0], "=")
				if len(keyAndValue) > 1 {
					os.Setenv(keyAndValue[0], keyAndValue[1])
				}
			}
		}
	}

	fmt.Println("==== ssh agent configured ===")

	fmt.Printf("\nSSH_AUTH_SOCK: %s", os.Getenv("SSH_AUTH_SOCK"))
	fmt.Printf("\nSSH_AGENT_PID: %s", os.Getenv("SSH_AGENT_PID"))

	sshAdd := exec.Command("ssh-add", sshKey)
	var out1 bytes.Buffer
	var stderr1 bytes.Buffer
	sshAdd.Stdout = &out1
	sshAdd.Stderr = &stderr1
	err = sshAdd.Run()
	if err != nil {
		fmt.Println("Executing: " + sshAdd.String())
		fmt.Println(fmt.Sprint(err) + ": " + stderr1.String())
		return
	}
	fmt.Println("Result: " + out1.String())

*/
