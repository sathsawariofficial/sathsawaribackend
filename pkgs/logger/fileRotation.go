package logger

import (
	"fmt"
	"os"
	"time"
)

type FileRotateWriter struct {
	CompleteFileName string
	fileName         string
	filePath         string
	Interval         time.Duration
	file             *os.File
	lastRotated      time.Time
	size             int
	MaxSize          int
}

func NewTimeRotationWriter(filename, filePath string, interval time.Duration, maxSize int) *FileRotateWriter {
	completePath := fmt.Sprintf("%v%v", filePath, filename)

	return &FileRotateWriter{
		fileName:         filename,
		filePath:         filePath,
		lastRotated:      time.Now(),
		CompleteFileName: completePath,
		Interval:         interval,
		MaxSize:          maxSize * MB,
	}
}

// rotates the file based on file size and time
func (w *FileRotateWriter) Write(output []byte) (int, error) {
	now := time.Now()

	// 1st condition is for time based rotation and 2nd is for size based rotation
	sizeRotationRequired := w.size+len(output) > w.MaxSize
	timeRotationRequired := now.Sub(w.lastRotated) > w.Interval

	if timeRotationRequired || sizeRotationRequired {
		if err := w.rotate(); err != nil {
			return 0, err
		}

		if timeRotationRequired {
			w.lastRotated = time.Now()
		}
	}

	if w.file == nil {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	n, err := w.file.Write(output)
	w.size += n

	return n, err
}

// renames old file and creates a new one
func (w *FileRotateWriter) rotate() error {
	if _, err := os.Stat(w.CompleteFileName); err == nil {
		// 2023-04-06T12-05-27.621_logger.log
		backupFilename := fmt.Sprintf("%v/%v_%v", w.filePath, time.Now().Format("2006-01-02T15-04-05.000"), w.fileName)
		if err := os.Rename(w.CompleteFileName, backupFilename); err != nil {
			return err
		}
	}

	file, err := os.Create(w.CompleteFileName)
	if err != nil {
		return err
	}

	w.file = file
	w.size = 0

	return nil
}
