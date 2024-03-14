package ssh

import (
	"bytes"
	"crypto"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"
	"github.com/koor-tech/genesis/pkg/models"
	"golang.org/x/crypto/ssh"
)

type Manager struct {
	logger *slog.Logger
}

func NewManager(logger *slog.Logger) *Manager {
	return &Manager{
		logger: logger,
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

func (m *Manager) RunAgent(sshPrivateKey string) error {
	sshAgent := exec.Command("ssh-agent", "-s")
	var out bytes.Buffer
	sshAgent.Stdout = &out

	if err := sshAgent.Run(); err != nil {
		m.logger.Error("error running ssh agent", "err", err)
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
	if err := sshAdd.Run(); err != nil {
		fmt.Println("Executing: " + sshAdd.String())
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}
	return nil
}
