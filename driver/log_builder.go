package driver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type jsonLogLine struct {
	Message       string    `json:"message"`
	ContainerId   string    `json:"container_id"`
	ContainerName string    `json:"container_name"`
	Tag           string    `json:"tag"`
	Host          string    `json:"host"`
	Timestamp     time.Time `json:"timestamp"`
}

func logMessage(lp *logPair, message []byte) error {
	lp.logLine.Message = string(message[:])
	lp.logLine.Timestamp = time.Now()

	data, err := json.Marshal(lp.logLine)
	if err != nil {
		return err
	}

	_, err = http.Post(lp.endpoint, "application/json", bytes.NewReader(data))

	return err
}
