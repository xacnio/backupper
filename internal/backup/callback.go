package backup

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (b *Backup) callCallback() error {
	callbackUrl, err := url.Parse(b.CallbackURL)
	if err != nil {
		return err
	}

	postData := make(map[string]interface{})
	postData["backup_date"] = b.StartedAt.Format(time.RFC3339)
	postData["backup_ts"] = b.StartedAt.Unix()
	postData["backup_id"] = b.stringID()
	postData["backup_name"] = b.Name
	postData["backup_source"] = b.Source.Type
	postData["backup_destination"] = b.Destination.Type
	postData["backup_destination_result"] = b.Destination.Result
	postData["backup_duration"] = time.Since(b.StartedAt).String()

	jsonData, _ := json.Marshal(postData)
	postDataBuffer := strings.NewReader(string(jsonData))

	httpClient := http.Client{}
	req, err := http.NewRequest("POST", callbackUrl.String(), postDataBuffer)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Backupper")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return err
	}

	return nil
}
