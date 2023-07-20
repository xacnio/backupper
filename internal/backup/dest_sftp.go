package backup

import (
	"github.com/xacnio/backupper/internal/utils"
	"github.com/xacnio/backupper/internal/utils/logger"
	"github.com/xacnio/backupper/pkg/sftp"
	"os"
	"path"
	"strings"
)

type DestinationSFTPInfo struct {
	Host           string  `json:"host"`
	Port           int     `json:"port"`
	User           string  `json:"user"`
	Pass           string  `json:"pass"`
	PrivateKeyFile string  `json:"privateKeyFile"`
	Passphrase     string  `json:"passphrase"`
	Target         string  `json:"target"`
	LimitByDate    *string `json:"limitByDate"`
	LimitByCount   *int    `json:"limitByCount"`
	LimitBySize    *int64  `json:"limitBySize"`
}

func (b *Backup) runDestinationSFTP() error {
	destination := b.Destination
	info := utils.ConvertToStruct[DestinationSFTPInfo](destination.Info)

	b.Destination.Result = DestinationResult{
		TotalUploadedFiles: 0,
		TotalUploadedSize:  0,
	}

	sftpConn := sftp.New(sftp.ConnConfig{
		Host:       info.Host,
		Port:       info.Port,
		User:       info.User,
		Pass:       info.Pass,
		PrivateKey: info.PrivateKeyFile,
		Passphrase: info.Passphrase,
	})

	err := sftpConn.Connect()
	if err != nil {
		logger.SFTP.Errorw("connection error", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "error", err)
		return err
	}

	defer sftpConn.Disconnect()

	logger.SFTP.Debugw("connection success", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port)

	// Create target folder
	_ = sftpConn.Client.MkdirAll(info.Target)

	// List files in ./tmp/{id}/
	tmpDir := "./tmp/" + b.stringID() + "/"
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		logger.FTP.Errorw("tmp directory error", "name", b.Name, "id", b.ID, "error", err)
		return err
	}

	// Upload files
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		f, err := os.Open(tmpDir + file.Name())
		if err != nil {
			logger.Main.Errorw("failed to backup because open file error", "name", b.Name, "id", b.ID)
			continue
		}
		fileBaseName := strings.TrimSuffix(file.Name(), path.Ext(file.Name()))
		fileExtension := path.Ext(file.Name())
		remoteFileName := fileBaseName + "-" + b.getFileTimeFormat() + fileExtension
		remoteFile, err := sftpConn.Client.Create(info.Target + "/" + remoteFileName)
		if err != nil {
			f.Close()
			logger.SFTP.Errorw("upload error", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "error", err)
			return err
		} else {
			_, err = remoteFile.ReadFrom(f)
			if err != nil {
				f.Close()
				remoteFile.Close()
				logger.SFTP.Errorw("upload error", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "error", err)
				return err
			}

			info2, err := remoteFile.Stat()
			if err == nil {
				b.Destination.Result.TotalUploadedFiles += 1
				b.Destination.Result.TotalUploadedSize += info2.Size()
			} else {
				logger.Main.Errorw("file info error", "name", b.Name, "id", b.ID, "error", err)
			}

			remoteFile.Close()
			f.Close()

			logger.SFTP.Debugw("upload success", "name", b.Name, "id", b.ID, "file", file.Name())
		}
	}

	if info.LimitByCount != nil && *info.LimitByCount > 0 {
		deleted, err := sftpConn.LimitByFileCount(info.Target, *info.LimitByCount)
		if err != nil {
			logger.SFTP.Errorw("limit by count error", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "error", err, "limit", *info.LimitByCount)
		} else {
			logger.SFTP.Infow("limit by count success", "name", b.Name, "id", b.ID, "limit", *info.LimitByCount, "deleted", deleted)
		}
	}

	if info.LimitBySize != nil && *info.LimitBySize > 0 {
		deleted, err := sftpConn.LimitByLength(info.Target, *info.LimitBySize)
		if err != nil {
			logger.SFTP.Errorw("limit by size error", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "error", err, "limit", *info.LimitBySize)
		} else {
			logger.SFTP.Infow("limit by size success", "name", b.Name, "id", b.ID, "limit", *info.LimitBySize, "deleted", deleted)
		}
	}

	if info.LimitByDate != nil {
		deleted, err := sftpConn.LimitByDate(info.Target, *info.LimitByDate)
		if err != nil {
			logger.SFTP.Errorw("limit by date error", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "error", err, "limit", *info.LimitByDate)
		} else {
			logger.SFTP.Infow("limit by date success", "name", b.Name, "id", b.ID, "limit", *info.LimitByDate, "deleted", deleted)
		}
	}
	return nil
}
