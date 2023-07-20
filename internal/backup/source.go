package backup

type SourceInfo struct {
	Type string      `json:"type"`
	Info interface{} `json:"info"`
}

func (b *Backup) runSource() error {
	source := b.Source
	switch source.Type {
	case "ftp":
		return b.runSourceFTP()
	case "sftp":
		return b.runSourceSFTP()
	}
	return nil
}
