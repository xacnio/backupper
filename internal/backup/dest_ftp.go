package backup

import (
	"github.com/xacnio/backupper/internal/utils"
	"github.com/xacnio/backupper/internal/utils/logger"
	"github.com/xacnio/backupper/pkg/ftp"
	"os"
	"path"
	"strings"
)

type DestinationFTPInfo struct {
	Host         string  `json:"host"`
	Port         int     `json:"port"`
	User         string  `json:"user"`
	Pass         string  `json:"pass"`
	Target       string  `json:"target"`
	LimitByDate  *string `json:"limitByDate"`
	LimitByCount *int    `json:"limitByCount"`
	LimitBySize  *uint64 `json:"limitBySize"`
}

func (b *Backup) runDestinationFTP() error {
	destination := b.Destination
	info := utils.ConvertToStruct[DestinationFTPInfo](destination.Info)

	b.Destination.Result = DestinationResult{
		TotalUploadedFiles: 0,
		TotalUploadedSize:  0,
	}

	conn := ftp.New(ftp.ConnConfig{
		Host: info.Host,
		Port: info.Port,
		User: info.User,
		Pass: info.Pass,
	})
	err := conn.Connect()
	if err != nil {
		logger.FTP.Errorw("connection error", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "error", err)
		return err
	}
	defer conn.Disconnect()

	// Create target folder
	folders := strings.Split(info.Target, "/")
	for i := 0; i < len(folders); i++ {
		folder := strings.Join(folders[:i+1], "/")
		_ = conn.MakeDir(folder)
	}

	// Go to target folder
	err = conn.ChangeDir(info.Target)
	if err != nil {
		logger.FTP.Errorw("change directory error", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "error", err)
		return err
	}

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
		err = conn.Stor(remoteFileName, f)
		if err != nil {
			f.Close()
			logger.FTP.Errorw("upload error", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "error", err)
			return err
		} else {
			logger.FTP.Debugw("upload success", "name", b.Name, "id", b.ID, "file", file.Name())

			fInfo, err := f.Stat()
			if err == nil {
				b.Destination.Result.TotalUploadedFiles++
				b.Destination.Result.TotalUploadedSize += fInfo.Size()
			}

			f.Close()
		}
	}

	if info.LimitByCount != nil && *info.LimitByCount > 0 {
		deleted, err := conn.LimitByFileCount(info.Target, *info.LimitByCount)
		if err != nil {
			logger.FTP.Errorw("limit by count error", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "error", err, "limit", *info.LimitByCount)
		} else {
			logger.FTP.Infow("limit by count success", "name", b.Name, "id", b.ID, "limit", *info.LimitByCount, "deleted", deleted)
		}
	}

	if info.LimitBySize != nil && *info.LimitBySize > 0 {
		deleted, err := conn.LimitByLength(info.Target, *info.LimitBySize)
		if err != nil {
			logger.FTP.Errorw("limit by size error", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "error", err, "limit", *info.LimitBySize)
		} else {
			logger.FTP.Infow("limit by size success", "name", b.Name, "id", b.ID, "limit", *info.LimitBySize, "deleted", deleted)
		}
	}

	if info.LimitByDate != nil {
		deleted, err := conn.LimitByDate(info.Target, *info.LimitByDate)
		if err != nil {
			logger.FTP.Errorw("limit by date error", "name", b.Name, "id", b.ID, "host", info.Host, "port", info.Port, "error", err, "limit", *info.LimitByDate)
		} else {
			logger.FTP.Infow("limit by date success", "name", b.Name, "id", b.ID, "limit", *info.LimitByDate, "deleted", deleted)
		}
	}
	return nil
}
