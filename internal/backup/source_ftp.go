package backup

import (
	"errors"
	"github.com/xacnio/backupper/internal/utils"
	"github.com/xacnio/backupper/internal/utils/logger"
	"github.com/xacnio/backupper/pkg/ftp"
	"os"
)

type SourceFTPInfo struct {
	Host      string   `json:"host"`
	Port      int      `json:"port"`
	User      string   `json:"user"`
	Pass      string   `json:"pass"`
	Downloads []string `json:"downloads"`
}

func (b *Backup) runSourceFTP() error {
	source := b.Source
	info := utils.ConvertToStruct[SourceFTPInfo](source.Info)

	if len(info.Downloads) == 0 {
		return errors.New("empty downloads")
	}

	ftpConn := ftp.New(ftp.ConnConfig{
		Host: info.Host,
		Port: info.Port,
		User: info.User,
		Pass: info.Pass,
	})

	err := ftpConn.Connect()
	if err != nil {
		logger.FTP.Errorw("connection error", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "error", err)
		return err
	} else {
		defer ftpConn.Disconnect()

		logger.SSH.Debugw("connection success", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port)

		// Tmp local directory
		tmpDir := "./tmp/" + b.stringID() + "/"
		err = os.MkdirAll(tmpDir, 0777)
		if err != nil {
			logger.Main.Errorw("tmp directory error", "name", b.Name, "id", b.ID, "error", err)
			return err
		}

		// Download files
		allErr := true
		for _, downloadFile := range info.Downloads {
			logger.FTP.Debugw("download started", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "file", downloadFile)

			err = ftpConn.Download(downloadFile, tmpDir+downloadFile)
			if err != nil {
				logger.FTP.Errorw("download error", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "error", err)
				continue
			} else {
				allErr = false
				logger.FTP.Debugw("download success", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "file", downloadFile)
			}
		}
		if allErr {
			return errors.New("all downloads failed")
		}
	}
	return nil
}
