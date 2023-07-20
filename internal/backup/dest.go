package backup

type DestinationResult struct {
	TotalUploadedFiles int64 `json:"totalUploadedFiles"`
	TotalUploadedSize  int64 `json:"totalUploadedSize"`
}

type DestinationInfo struct {
	Type              string            `json:"type"`
	DeleteAfterUpload *bool             `json:"deleteAfterUpload"`
	Info              interface{}       `json:"info"`
	Result            DestinationResult `json:"-"`
}

func (b *Backup) runDestination() error {
	dest := b.Destination
	switch dest.Type {
	case "sftp":
		return b.runDestinationSFTP()
	case "ftp":
		return b.runDestinationFTP()
	}
	return nil
}
