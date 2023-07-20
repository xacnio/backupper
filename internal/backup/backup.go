package backup

import (
	"github.com/go-co-op/gocron"
	"github.com/xacnio/backupper/internal/config"
	"github.com/xacnio/backupper/internal/utils/logger"
	"os"
	"strconv"
	"time"
)

type Backup struct {
	ID             int64           `json:"-"`
	Name           string          `json:"name"`
	Source         SourceInfo      `json:"source"`
	Destination    DestinationInfo `json:"destination"`
	CronExpression string          `json:"cronExpr"`
	StartedAt      time.Time       `json:"-"`
	CallbackURL    string          `json:"callbackUrl"`
	DeleteLocal    *bool           `json:"deleteLocal"`
	Job            *gocron.Job     `json:"-"`
}

func (b *Backup) clear() error {

	tmp := "./tmp/" + b.stringID() + "/"
	err := os.RemoveAll(tmp)
	if err != nil {
		return err
	}
	return nil
}

func (b *Backup) CreateFunc() func() {
	return func() {
		b.StartedAt = time.Now()
		b.ID = b.StartedAt.UnixNano()

		logger.Main.Infow("backup started", "name", b.Name, "id", b.ID)

		var err error
		err = b.runSource()
		if err != nil {
			logger.Main.Errorw("source error", "name", b.Name, "id", b.ID, "error", err)
			return
		} else {
			logger.Main.Debugw("source success", "name", b.Name, "id", b.ID)
		}

		err = b.runDestination()
		if err != nil {
			logger.Main.Errorw("backup error", "name", b.Name, "id", b.ID, "error", err)
		} else {
			logger.Main.Infow("backup success", "name", b.Name, "id", b.ID)
		}

		if b.CallbackURL == "" {
			logger.Main.Debugw("callback none", "name", b.Name, "id", b.ID)
		} else {
			err = b.callCallback()
			if err != nil {
				logger.Main.Errorw("callback error", "name", b.Name, "id", b.ID, "error", err)
			} else {
				logger.Main.Debugw("callback success", "name", b.Name, "id", b.ID)
			}
		}

		if b.DeleteLocal == nil || *b.DeleteLocal == true {
			err = b.clear()
			if err != nil {
				logger.Main.Errorw("tmp clear error", "name", b.Name, "id", b.ID, "error", err)
			} else {
				logger.Main.Debugw("tmp cleared", "name", b.Name, "id", b.ID)
			}
		} else {
			logger.Main.Debugw("deleteLocal false", "name", b.Name, "id", b.ID)
		}

		logger.Main.Infow("backup finished", "name", b.Name, "id", b.ID)
	}
}

func (b *Backup) getFileTimeFormat() string {
	return b.StartedAt.Format(config.Get().DateFormat)
}

func (b *Backup) stringID() string {
	id := strconv.FormatInt(b.ID, 10)
	return id
}
