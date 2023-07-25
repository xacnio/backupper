package sftp

import (
	"fmt"
	"github.com/pkg/sftp"
	"github.com/xacnio/backupper/internal/utils"
	"github.com/xacnio/backupper/internal/utils/logger"
	ssh2 "golang.org/x/crypto/ssh"
	"os"
	"path"
	"sort"
	"time"
)

type SFTP struct {
	Connected bool
	SSHClient *ssh2.Client
	Client    *sftp.Client
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

func New(c ConnConfig) *SFTP {
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
			logger.SFTP.Error("Unable to read private key: %v", err)
		} else {
			if c.Passphrase != "" {
				signer, err := ssh2.ParsePrivateKeyWithPassphrase(key, []byte(c.Passphrase))
				if err != nil {
					logger.SFTP.Errorw("unable to parse private key", "host", c.Host, "port", c.Port, "error", err)
				} else {
					authInfo.Auth = []ssh2.AuthMethod{
						ssh2.PublicKeys(signer),
					}
				}
			} else {
				signer, err := ssh2.ParsePrivateKey(key)
				if err != nil {
					logger.SFTP.Errorw("unable to parse private key", "host", c.Host, "port", c.Port, "error", err)
				} else {
					authInfo.Auth = []ssh2.AuthMethod{
						ssh2.PublicKeys(signer),
					}
				}
			}
		}
	}
	return &SFTP{
		Config: c,
		Auth:   authInfo,
	}
}

func (f *SFTP) Disconnect() error {
	err := f.Client.Close()
	if err != nil {
		return err
	}
	err = f.SSHClient.Close()
	if err != nil {
		return err
	}
	f.Connected = false
	logger.SFTP.Debugw("disconnected", "host", f.Config.Host, "port", f.Config.Port)
	return nil
}

func (f *SFTP) Connect() error {
	var err error

	c := f.Config

	f.SSHClient, err = ssh2.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port), &f.Auth)
	if err != nil {
		return err
	}

	f.Client, err = sftp.NewClient(f.SSHClient)
	if err != nil {
		return err
	}

	f.Connected = true
	logger.SFTP.Debugw("connected", "host", f.Config.Host, "port", f.Config.Port)
	return nil
}

func (f *SFTP) DownloadFile(remotePath string, localPath string) error {
	remoteFile, err := f.Client.Open(remotePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	localFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	_, err = remoteFile.WriteTo(localFile)
	if err != nil {
		return err
	}
	return nil
}

func (f *SFTP) LimitByLength(remoteFolder string, limitBytes int64) ([]string, error) {
	entries, err := f.Client.ReadDir(remoteFolder)
	if err != nil {
		return []string{}, err
	}

	var fileEntries []os.FileInfo
	var totalLength int64 = 0
	for i := range entries {
		if entries[i].IsDir() {
			continue
		}
		totalLength += entries[i].Size()
		fileEntries = append(fileEntries, entries[i])
	}
	sort.Slice(fileEntries, func(i, j int) bool {
		return fileEntries[i].ModTime().Before(fileEntries[j].ModTime())
	})
	if totalLength > limitBytes {
		var deletedFiles []string
		diff := totalLength
		diff -= limitBytes
		for i := range fileEntries {
			if diff <= 0 {
				break
			}
			fileSize := fileEntries[i].Size()
			diff -= fileSize
			err = f.Client.Remove(path.Join(remoteFolder, fileEntries[i].Name()))
			if err == nil {
				deletedFiles = append(deletedFiles, fileEntries[i].Name())
			}
		}
		return deletedFiles, nil
	}
	return []string{}, nil
}

func (f *SFTP) LimitByFileCount(remoteFolder string, limit int) ([]string, error) {
	entries, err := f.Client.ReadDir(remoteFolder)
	if err != nil {
		return []string{}, err
	}

	var fileEntries []os.FileInfo
	for i := range entries {
		if entries[i].IsDir() {
			continue
		}
		fileEntries = append(fileEntries, entries[i])
	}
	sort.Slice(fileEntries, func(i, j int) bool {
		return fileEntries[i].ModTime().Before(fileEntries[j].ModTime())
	})
	if len(fileEntries) > limit {
		var deletedFiles []string
		diff := len(fileEntries) - limit
		for i := 0; i < diff; i++ {
			err = f.Client.Remove(path.Join(remoteFolder, fileEntries[i].Name()))
			if err == nil {
				deletedFiles = append(deletedFiles, fileEntries[i].Name())
			}
		}
	}
	return []string{}, nil
}

func (f *SFTP) LimitByDate(remoteFolder string, beforeDurationPattern string) ([]string, error) {
	beforeTime, ok := utils.ParseDurationPattern(beforeDurationPattern, true)
	if !ok {
		return []string{}, fmt.Errorf("invalid duration pattern")
	}

	entries, err := f.Client.ReadDir(remoteFolder)
	if err != nil {
		return []string{}, err
	}

	var deletedFiles []string
	for i := range entries {
		if entries[i].IsDir() {
			continue
		}
		if entries[i].ModTime().Before(beforeTime) {
			err = f.Client.Remove(path.Join(remoteFolder, entries[i].Name()))
			if err == nil {
				deletedFiles = append(deletedFiles, entries[i].Name())
			}
		}
	}
	return deletedFiles, nil
}
