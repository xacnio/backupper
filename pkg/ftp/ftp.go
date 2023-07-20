package ftp

import (
	"fmt"
	"github.com/jlaffaye/ftp"
	"github.com/xacnio/backupper/internal/utils"
	"github.com/xacnio/backupper/internal/utils/logger"
	"io"
	"os"
	"path"
	"sort"
	"time"
)

type FTP struct {
	Connected bool
	*ftp.ServerConn
	Host string
	Port int
	User string
	Pass string
}

type ConnConfig struct {
	Host string
	Port int
	User string
	Pass string
}

func New(c ConnConfig) *FTP {
	return &FTP{
		Host: c.Host,
		Port: c.Port,
		User: c.User,
		Pass: c.Pass,
	}
}

func (f *FTP) Disconnect() error {
	err := f.ServerConn.Quit()
	if err != nil {
		return err
	}
	f.Connected = false
	logger.FTP.Debugw("disconnected", "host", f.Host, "port", f.Port)
	return nil
}

func (f *FTP) Connect() error {
	var err error
	f.ServerConn, err = ftp.Dial(fmt.Sprintf("%s:%d", f.Host, f.Port), ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return err
	}

	err = f.ServerConn.Login(f.User, f.Pass)
	if err != nil {
		return err
	}
	f.Connected = true
	logger.FTP.Debugw("connected", "host", f.Host, "port", f.Port)
	return nil
}

func (f *FTP) Download(remotePath string, localPath string) error {
	res, err := f.ServerConn.Retr(remotePath)
	if err != nil {
		return err
	}
	defer res.Close()

	outFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, res)
	if err != nil {
		return err
	}
	return nil
}

func (f *FTP) LimitByFileCount(remoteFolder string, limit int) ([]string, error) {
	entries, err := f.ServerConn.List(remoteFolder)
	if err != nil {
		return []string{}, err
	}

	var fileEntries []*ftp.Entry
	for i := range entries {
		if entries[i].Type == ftp.EntryTypeFolder {
			continue
		}
		fileEntries = append(fileEntries, entries[i])
	}
	sort.Slice(fileEntries, func(i, j int) bool {
		return fileEntries[i].Time.Before(fileEntries[j].Time)
	})
	if len(fileEntries) > limit {
		var deletedFiles []string
		diff := len(fileEntries) - limit
		for i := 0; i < diff; i++ {
			err = f.ServerConn.Delete(path.Join(remoteFolder, fileEntries[i].Name))
			if err == nil {
				deletedFiles = append(deletedFiles, fileEntries[i].Name)
			}
		}
		return deletedFiles, nil
	}
	return []string{}, nil
}

func (f *FTP) LimitByLength(remoteFolder string, limitBytes uint64) ([]string, error) {
	entries, err := f.ServerConn.List(remoteFolder)
	if err != nil {
		return []string{}, err
	}

	var fileEntries []*ftp.Entry
	var totalLength uint64 = 0
	for i := range entries {
		if entries[i].Type == ftp.EntryTypeFolder {
			continue
		}
		totalLength += entries[i].Size
		fileEntries = append(fileEntries, entries[i])
	}
	sort.Slice(fileEntries, func(i, j int) bool {
		return fileEntries[i].Time.Before(fileEntries[j].Time)
	})
	if totalLength > limitBytes {
		var deletedFiles []string
		diff := totalLength - limitBytes
		for i := range fileEntries {
			if diff <= 0 {
				break
			}
			diff -= fileEntries[i].Size
			err = f.ServerConn.Delete(path.Join(remoteFolder, fileEntries[i].Name))
			if err == nil {
				deletedFiles = append(deletedFiles, fileEntries[i].Name)
			}
		}
		return deletedFiles, nil
	}
	return []string{}, nil
}

func (f *FTP) LimitByDate(remoteFolder string, beforeDurationPattern string) ([]string, error) {
	beforeTime, ok := utils.ParseDurationPattern(beforeDurationPattern, true)
	if !ok {
		return []string{}, fmt.Errorf("invalid duration pattern")
	}

	entries, err := f.ServerConn.List(remoteFolder)
	if err != nil {
		return []string{}, err
	}

	var deleteFiles []string
	for i := range entries {
		if entries[i].Type == ftp.EntryTypeFolder {
			continue
		}
		if entries[i].Time.Before(beforeTime) {
			err = f.ServerConn.Delete(path.Join(remoteFolder, entries[i].Name))
			if err == nil {
				deleteFiles = append(deleteFiles, entries[i].Name)
			}
		}
	}
	return deleteFiles, nil
}
