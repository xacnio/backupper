package backup

import (
	"bytes"
	"fmt"
	"github.com/xacnio/backupper/internal/utils"
	"github.com/xacnio/backupper/internal/utils/logger"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
)

type DestinationTelegramInfo struct {
	Token  string `json:"token"`
	ChatID string `json:"chatID"`
}

func (b *Backup) runDestinationTelegramBot() error {
	destination := b.Destination
	info := utils.ConvertToStruct[DestinationTelegramInfo](destination.Info)

	b.Destination.Result = DestinationResult{
		TotalUploadedFiles: 0,
		TotalUploadedSize:  0,
	}

	// List files in ./tmp/{id}/
	tmpDir := "./tmp/" + b.stringID() + "/"
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		logger.FTP.Errorw("tmp directory error", "name", b.Name, "id", b.ID, "error", err)
		return err
	}

	// Loop through files
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		f, err := os.Open(tmpDir + file.Name())
		if err != nil {
			logger.Main.Errorw("failed to backup because open file error", "name", b.Name, "id", b.ID)
			continue
		}
		uploadProc := func() error {
			defer f.Close()

			// Create request body
			var requestBody bytes.Buffer
			writer := multipart.NewWriter(&requestBody)
			writer.WriteField("chat_id", info.ChatID)
			part, err := writer.CreateFormFile("document", path.Base(f.Name()))
			if err != nil {
				logger.Main.Errorw("failed to backup because create form file error", "name", b.Name, "id", b.ID)
				return err
			}
			_, err = io.Copy(part, f)
			if err != nil {
				logger.Main.Errorw("failed to backup because copy file error", "name", b.Name, "id", b.ID)
				return err
			}
			writer.Close()

			// Create the HTTP POST request
			url := fmt.Sprintf("https://api.telegram.org/bot%s/sendDocument", info.Token)
			request, err := http.NewRequest("POST", url, &requestBody)
			if err != nil {
				logger.Main.Errorw("failed to backup because create request error", "name", b.Name, "id", b.ID)
				return err
			}
			// Set the Content-Type header
			request.Header.Set("Content-Type", writer.FormDataContentType())

			logger.TgBot.Debugw("telegram bot upload start", "name", b.Name, "id", b.ID, "file", file.Name())

			// Perform the HTTP request
			client := &http.Client{}
			response, err := client.Do(request)
			if err != nil {
				logger.Main.Errorw("failed to backup because send request error", "name", b.Name, "id", b.ID)
				return err
			}
			defer response.Body.Close()
			// body to string
			body, _ := io.ReadAll(response.Body)

			if response.StatusCode != 200 {
				logger.TgBot.Errorw("failed to backup because telegram bot error", "name", b.Name, "id", b.ID, "status", response.Status, "body", body)
			} else {
				info2, _ := f.Stat()
				b.Destination.Result.TotalUploadedFiles++
				b.Destination.Result.TotalUploadedSize += info2.Size()
				logger.TgBot.Debugw("telegram bot upload success", "name", b.Name, "id", b.ID, "file", file.Name(), "size", info2.Size())
			}
			return nil
		}
		_ = uploadProc()
	}

	return nil
}
