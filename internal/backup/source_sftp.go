package backup

import (
	"errors"
	"github.com/xacnio/backupper/internal/utils"
	"github.com/xacnio/backupper/internal/utils/logger"
	"github.com/xacnio/backupper/pkg/sftp"
	"github.com/xacnio/backupper/pkg/ssh"
	"os"
	"strconv"
	"strings"
)

type SourceSFTPInfo struct {
	Host           string                  `json:"host"`
	Port           int                     `json:"port"`
	User           string                  `json:"user"`
	Pass           string                  `json:"pass"`
	Variables      *map[string]interface{} `json:"variables"`
	BeforeCommands []string                `json:"beforeCommands"`
	Downloads      []string                `json:"downloads"`
	AfterCommands  []string                `json:"afterCommands"`
	PrivateKeyFile string                  `json:"privateKeyFile"`
	Passphrase     string                  `json:"passphrase"`
}

func (b *Backup) runSourceSFTP() error {
	source := b.Source
	info := utils.ConvertToStruct[SourceSFTPInfo](source.Info)

	sftpConn := sftp.New(sftp.ConnConfig{
		Host:       info.Host,
		Port:       info.Port,
		User:       info.User,
		Pass:       info.Pass,
		PrivateKey: info.PrivateKeyFile,
		Passphrase: info.Passphrase,
	})

	sshConn := ssh.New(ssh.ConnConfig{
		Host:       info.Host,
		Port:       info.Port,
		User:       info.User,
		Pass:       info.Pass,
		PrivateKey: info.PrivateKeyFile,
		Passphrase: info.Passphrase,
	})
	err := sshConn.Connect()
	if err != nil {
		logger.SSH.Errorw("connection error", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "error", err)
		return err
	} else {
		defer sshConn.Disconnect()

		logger.SSH.Debugw("connection success", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port)

		if len(info.BeforeCommands) > 0 {
			nameEscaped := strings.ReplaceAll(b.Name, "\"", "\\\"")
			commands := []string{
				"BACKUP_ID=" + b.stringID(),
				"BACKUP_NAME=\"" + nameEscaped + "\"",
				"mkdir -p /tmp/backupper/$BACKUP_ID",
			}
			if info.Variables != nil {
				for k, v := range *info.Variables {
					value := ""
					switch v.(type) {
					case string:
						v = strings.ReplaceAll(v.(string), "\"", "\\\"")
						value = "\"" + v.(string) + "\""
					case int, int8, int16, int32, int64:
						value = strconv.Itoa(v.(int))
					case float32, float64:
						value = strconv.FormatFloat(v.(float64), 'f', -1, 64)
					case bool:
						value = strconv.FormatBool(v.(bool))
					default:
						continue
					}
					commands = append(commands, k+"="+value)
				}
			}
			commands = append(commands, info.BeforeCommands...)

			combinedOutput, err := sshConn.RunCommands(commands)
			if err != nil {
				logger.SSH.Errorw("before commands error", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "error", err)
				return err
			} else {
				logger.SSH.Debugw("before commands success", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "output", combinedOutput, "commands", commands)
			}
		}

		// Tmp local directory
		tmpDir := "./tmp/" + b.stringID() + "/"
		err = os.MkdirAll(tmpDir, 0777)
		if err != nil {
			logger.SSH.Errorw("tmp directory error", "name", b.Name, "id", b.ID, "error", err)
			return err
		}

		// SFTP session
		err = sftpConn.Connect()
		if err != nil {
			logger.SFTP.Errorw("connection error", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "error", err)
			return err
		}
		defer sftpConn.Disconnect()

		remoteFolder := "/tmp/backupper/" + b.stringID() + "/"
		allErr := true
		for _, downloadFile := range info.Downloads {
			remotePath := remoteFolder + downloadFile
			if strings.HasPrefix(downloadFile, "/") {
				remotePath = downloadFile
			}
			logger.SFTP.Debugw("download start", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "remotePath", remotePath, "localPath", tmpDir+downloadFile)
			err = sftpConn.DownloadFile(remotePath, tmpDir+downloadFile)
			if err != nil {
				logger.SFTP.Errorw("download error", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "error", err)
				continue
			} else {
				allErr = false
				logger.SFTP.Debugw("download success", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "remotePath", remotePath, "localPath", tmpDir+downloadFile)
			}
		}
		if allErr {
			return errors.New("download error")
		}

		if len(info.AfterCommands) > 0 {
			nameEscaped := strings.ReplaceAll(b.Name, "\"", "\\\"")
			commands := []string{
				"BACKUP_ID=" + b.stringID(),
				"BACKUP_NAME=\"" + nameEscaped + "\"",
			}
			commands = append(commands, info.AfterCommands...)

			combinedOutput, err := sshConn.RunCommands(commands)
			if err != nil {
				logger.SSH.Errorw("after commands error", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "error", err)
				return err
			} else {
				logger.SSH.Debugw("after commands success", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "output", combinedOutput, "commands", commands)
			}
		}
	}
	return nil
}
