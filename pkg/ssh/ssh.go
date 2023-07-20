package ssh

import (
	"bytes"
	"fmt"
	"github.com/xacnio/backupper/internal/utils/logger"
	ssh2 "golang.org/x/crypto/ssh"
	"log"
	"os"
	"time"
)

type SSH struct {
	Connected bool
	Client    *ssh2.Client
	Auth      ssh2.ClientConfig
	Config    ConnConfig
}

type ConnConfig struct {
	Host       string
	Port       int
	User       string
	Pass       string
	PrivateKey string
	Passphrase string
}

func New(c ConnConfig) *SSH {
	authInfo := ssh2.ClientConfig{
		User: c.User,
		Auth: []ssh2.AuthMethod{
			ssh2.Password(c.Pass),
		},
		Timeout:         5 * time.Second,
		HostKeyCallback: ssh2.InsecureIgnoreHostKey(),
	}
	if c.PrivateKey != "" {
		key, err := os.ReadFile(c.PrivateKey)
		if err != nil {
			log.Fatalf("Unable to read private key: %v", err)
		} else {
			if c.Passphrase != "" {
				signer, err := ssh2.ParsePrivateKeyWithPassphrase(key, []byte(c.Passphrase))
				if err != nil {
					logger.SSH.Errorw("unable to parse private key", "host", c.Host, "port", c.Port, "error", err)
				} else {
					authInfo.Auth = []ssh2.AuthMethod{
						ssh2.PublicKeys(signer),
					}
				}
			} else {
				signer, err := ssh2.ParsePrivateKey(key)
				if err != nil {
					logger.SSH.Errorw("unable to parse private key", "host", c.Host, "port", c.Port, "error", err)
				} else {
					authInfo.Auth = []ssh2.AuthMethod{
						ssh2.PublicKeys(signer),
					}
				}
			}
		}
	}
	return &SSH{
		Config: c,
		Auth:   authInfo,
	}
}

func (f *SSH) Disconnect() error {
	err := f.Client.Close()
	if err != nil {
		return err
	}
	f.Connected = false
	logger.SSH.Debugw("disconnected", "host", f.Config.Host, "port", f.Config.Port)
	return nil
}

func (f *SSH) Connect() error {
	var err error

	c := f.Config

	f.Client, err = ssh2.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port), &f.Auth)
	if err != nil {
		return err
	}

	f.Connected = true
	logger.SSH.Debugw("connected", "host", f.Config.Host, "port", f.Config.Port)
	return nil
}

func (f *SSH) RunCommands(commands []string) (string, error) {
	if len(commands) == 0 {
		return "", nil
	}
	if !f.Connected {
		return "", fmt.Errorf("not connected to ssh")
	}
	session, err := f.Client.NewSession()
	if err != nil {
		return "", err
	} else {
		defer session.Close()

		var bf bytes.Buffer
		session.Stdout = &bf
		session.Stderr = &bf

		stdin, err := session.StdinPipe()
		if err != nil {
			return "", err
		}
		err = session.Shell()
		if err != nil {
			return "", err
		} else {
			for _, command := range commands {
				stdin.Write([]byte(command + "\n"))
			}

			stdin.Write([]byte("exit\n"))

			session.Wait()
			return bf.String(), nil
		}
	}
}
