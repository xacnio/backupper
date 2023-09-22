### (IN DEVELOPMENT)

# Backupper
The tool is for backuping file from FTP/SFTP source and save it another FTP/SFTP backup destination.

# Features
- Schedule backuping with CRON expression in config file
- Download files from FTP/SFTP sources
- Executing commands on source server (SFTP only)
- Upload files to FTP/SFTP destinations
- Upload files to Telegram with Bot API (max 50 MB files)
- Limit the file count in target folder (if limit is reached, oldest backups will be deleted)
- Limit the total file size in target folder (if limit is reached, oldest backups will be deleted)
- Limit the target folder by duration (oldest backups than date will be deleted)

# Config
Main config file is `config.json` in the root folder of the project.
Also you can use [`config.example.json`](config.example.json) as a template.

### Main Structure
```json
{
  "timezone": "Europe/Istanbul",
  "dateFormat": "2006-01-02__15-04-05",
  "backups": []
}
```

| Key        | Description                                                                                                   | Type   |
|------------|---------------------------------------------------------------------------------------------------------------|--------|
| timezone   | [TZ identifier](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones)                                 | string |
| dateFormat | the format of the date to be added to the file name ([Golang time format](https://go.dev/src/time/format.go)) | string |
| logLevel   | Log level (debug, info, warn, error, dpanic, panic, fatal)                                                    | string |
| backups    | Backup schedules                                                                                              | array  |

## Backup Schedule Structure
```json
{
  "name": "backup-title",
  "cronExpr": "*/60 * * * * *",
  "callbackUrl": "http://example.com/callback",
  "deleteLocal": true,
  "source": {
    "type": "sftp",
    "info": {
      "host": "127.0.0.1",
      "port": 22,
      "user": "root",
      "pass": "",
      "privateKeyFile": "~/.ssh/custom_id_rsa",
      "privateKeyPass": "",
      "variables": {
        "foo": "bar",
        "bar": 123,
        "baz": true
      },
      "beforeCommands": [
        "echo $foo > /tmp/backupper/$BACKUP_ID/foo.txt",
        "echo $bar > /tmp/backupper/$BACKUP_ID/bar.txt",
        "echo $baz > /tmp/backupper/$BACKUP_ID/baz.txt",
        "cd /opt/foo/bar",
        "zip -r /tmp/backupper/$BACKUP_ID/foo_backup.zip ."
      ],
      "downloads": [
        "foo_backup.zip",
        "foo.txt",
        "bar.txt",
        "baz.txt"
      ],
      "afterCommands": [
        "rm -rf /tmp/backupper/$BACKUP_ID/"
      ]
    }
  },
  "destination": {
    "type": "ftp",
    "deleteAfterUpload": true,
    "info": {
      "host": "backup-ftp.example.com",
      "port": 21,
      "user": "root",
      "pass": "p4ssw0rd",
      "target": "/up/backups/foo/",
      "limitByCount": 3,
      "LimitBySize": 1073741824,
      "limitByDate": "2 DAYS"
    }
  }
}
```

| Key         | Description                                                                       | Type   |
|-------------|-----------------------------------------------------------------------------------|--------|
| name        | Name of the backup schedule                                                       | string |
| cronExpr    | [CRON expression](https://en.wikipedia.org/wiki/Cron) (also supports cronSeconds) | string |
| callbackUrl | Callback URL to be called after backup process is completed                       | string |
| deleteLocal | Delete local files after upload process is completed                              | bool   |
| source      | Source server information                                                         | object |
| destination | Destination server information                                                    | object |

## Source 
| Key        | Description                                                                         | Type   |
|------------|-------------------------------------------------------------------------------------|--------|
| type       | Source server type (ftp/sftp)                                                       | string |
| info       | Source server information                                                           | object |

### Source Info (FTP)
| Key       | Description                                                          | Type   |
|-----------|----------------------------------------------------------------------|--------|
| host      | FTP server host                                                      | string |
| port      | FTP server port                                                      | int    |
| user      | FTP server username                                                  | string |
| pass      | FTP server password                                                  | string |
| downloads | Files to be downloaded from FTP server (only files, not directories) | array  |

### Source Info (SFTP)
| Key            | Description                                                           | Type   |
|----------------|-----------------------------------------------------------------------|--------|
| host           | SFPT server host                                                      | string |
| port           | SFTP server port                                                      | int    |
| user           | SFTP server username                                                  | string |
| pass           | SFTP server password                                                  | string |
| privateKeyFile | Private key file path                                                 | string |
| passphrase     | Private key passphrase                                                | string |
| variables      | Custom variables to be used in SSH commands                           | object |
| beforeCommands | SSH Commands to be executed before download process                   | array  |
| downloads      | Files to be downloaded from SFTP server (only files, not directories) | array  |
| afterCommands  | SSH Commands to be executed after download process                    | array  |

#### SFTP - SSH Command Variables
| Variable     | Description                                                                   | Type   |
|--------------|-------------------------------------------------------------------------------|--------|
| $BACKUP_ID   | Unique ID of the backup process (generated by the tool) (Nano unix timestamp) | string |
| $BACKUP_NAME | Name of the backup schedule                                                   | string |

## Destination
| Key               | Description                                     | Type   |
|-------------------|-------------------------------------------------|--------|
| type              | Destination server type (ftp/sftp/telegram_bot) | string |
| deleteAfterUpload | Delete files after upload process is completed  | bool   |
| info              | Destination server information                  | object |

### Destination Info (Telegram with Bot API) (max 50 MB files)
| Key          | Description                                                  | Type   |
|--------------|--------------------------------------------------------------|--------|
| token        | Telegram bot token from [@BotFather](https://t.me/BotFather) | string |
| chatID       | Telegram chat ID (channel/group) or public username          | string |

### Destination Info (FTP)
| Key          | Description                                           | Type   |
|--------------|-------------------------------------------------------|--------|
| host         | FTP server host                                       | string |
| port         | FTP server port                                       | int    |
| user         | FTP server username                                   | string |
| pass         | FTP server password                                   | string |
| target       | Target folder on FTP server                           | string |
| limitByCount | Limit the file count in target folder                 | int    |
| limitBySize  | Limit the total file size in target folder (bytes)    | int    |
| limitByDate  | Limit the target folder by duration (duration format) | string |

### Destination Info (SFTP)
| Key            | Description                                           | Type   |
|----------------|-------------------------------------------------------|--------|
| host           | SFTP server host                                      | string |
| port           | SFTP server port                                      | int    |
| user           | SFTP server username                                  | string |
| pass           | SFTP server password                                  | string |
| privateKeyFile | Private key file path                                 | string |
| passphrase     | Private key passphrase                                | string |
| target         | Target folder on SFTP server                          | string |
| limitByCount   | Limit the file count in target folder                 | int    |
| limitBySize    | Limit the total file size in target folder (bytes)    | int    |
| limitByDate    | Limit the target folder by duration (duration format) | string |

#### limitByDate - Duration Format
| Format                | Date Range                    |
|-----------------------|-------------------------------|
| 5 MINUTES             | All files before 5 minutes    |
| 1 HOUR                | All files before 1 hour       |
| 2 DAYS                | All files before 2 days       |
| 3 WEEKS               | All files before 3 weeks      |
| 4 MONTHS              | All files before 4 months     |
| 5 YEARS               | All files before 5 years      |

Multiple durations can be used together. (e.g. 1 HOUR 30 MINUTES)

## Callback Post Data
```json
{
  "backup_date": "2023-07-20T15:27:50+03:00",
  "backup_destination": "sftp",
  "backup_destination_result": {
    "totalUploadedFiles": 1,
    "totalUploadedSize": 5820073
  },
  "backup_duration": "19.7989235s",
  "backup_id": "1689856070674044500",
  "backup_name": "test-backup",
  "backup_source": "sftp",
  "backup_ts": 1689856070
}
```

| Key                       | Description                                                                   | Type   |
|---------------------------|-------------------------------------------------------------------------------|--------|
| backup_date               | Backup date  (RFC3339)                                                        | string |
| backup_destination        | Destination server type (ftp/sftp)                                            | string |
| backup_destination_result | Destination server upload result                                              | object |
| backup_duration           | Backup duration (time.Duration string)                                        | string |
| backup_id                 | Unique ID of the backup process (generated by the tool) (Nano unix timestamp) | int    |
| backup_name               | Name of the backup schedule                                                   | string |
| backup_source             | Source server type (ftp/sftp)                                                 | string |
| backup_ts                 | Backup timestamp (Unix seconds)                                               | int    |

# Used Modules
- [go-co-op/gocron](https://pkg.go.dev/github.com/go-co-op/gocron)
- [jlaffaye/ftp](https://pkg.go.dev/github.com/jlaffaye/ftp)
- [uber/zap](https://pkg.go.dev/go.uber.org/zap)