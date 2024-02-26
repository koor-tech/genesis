package ssh

import (
	"bytes"
	"crypto"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/google/uuid"
	"github.com/koor-tech/genesis/pkg/models"
	"golang.org/x/crypto/ssh"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

type Manager struct {
	logger *slog.Logger
}

func NewManager() *Manager {
	return &Manager{
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
}

func (m *Manager) GenerateKey() (*models.Ssh, error) {
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

	return &models.Ssh{
		ID:         uuid.New(),
		PublicKey:  publicKeyString,
		PrivateKey: privateKeyString,
	}, nil
}

func (m *Manager) RunAddgentRunAgent() {

}

func (m *Manager) RunAgent(sshPrivateKey string) error {
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
	var sshOut bytes.Buffer
	var stderr bytes.Buffer
	sshAdd.Stdout = &sshOut
	sshAdd.Stderr = &stderr
	err = sshAdd.Run()
	if err != nil {
		fmt.Println("Executing: " + sshAdd.String())
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}
	return nil
}
