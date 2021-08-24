package driver

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"path"
	"sync"
	"syscall"

	"github.com/containerd/fifo"
	"github.com/docker/docker/api/types/plugins/logdriver"
	"github.com/docker/docker/daemon/logger"
	"github.com/docker/docker/daemon/logger/loggerutils"
	protoio "github.com/gogo/protobuf/io"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Driver struct {
	mu   sync.Mutex
	logs map[string]*logPair
}

type logPair struct {
	active   bool
	file     string
	info     logger.Info
	logLine  jsonLogLine
	stream   io.ReadCloser
	endpoint string
}

func NewDriver() *Driver {
	return &Driver{
		logs: make(map[string]*logPair),
	}
}

func (d *Driver) StartLogging(file string, logCtx logger.Info) error {
	d.mu.Lock()
	if _, exists := d.logs[path.Base(file)]; exists {
		d.mu.Unlock()
		return fmt.Errorf("logger for %q already exists", file)
	}
	d.mu.Unlock()

	logrus.WithField("id", logCtx.ContainerID).WithField("file", file).Info("Start logging")
	stream, err := fifo.OpenFifo(context.Background(), file, syscall.O_RDONLY, 0700)
	if err != nil {
		return errors.Wrapf(err, "error opening logger fifo: %q", file)
	}

	tag, err := loggerutils.ParseLogTag(logCtx, loggerutils.DefaultTemplate)
	if err != nil {
		return err
	}

	hostname, err := logCtx.Hostname()
	if err != nil {
		return err
	}

	logLine := jsonLogLine{
		ContainerId:   logCtx.FullID(),
		ContainerName: logCtx.Name(),
		Tag:           tag,
		Host:          hostname,
	}

	endpoint, ok := logCtx.Config["endpoint"]
	if !ok {
		return errors.New("enpoint parameter is required")
	}

	lp := &logPair{true, file, logCtx, logLine, stream, endpoint}

	d.mu.Lock()
	d.logs[path.Base(file)] = lp
	d.mu.Unlock()

	go consumeLog(lp)
	return nil
}

func (d *Driver) StopLogging(file string) error {
	d.mu.Lock()
	lp, ok := d.logs[path.Base(file)]
	if ok {
		lp.active = false
		delete(d.logs, path.Base(file))
	}
	d.mu.Unlock()
	return nil
}

func shutdownLogPair(lp *logPair) {
	if lp.stream != nil {
		lp.stream.Close()
	}

	lp.active = false
}

func consumeLog(lp *logPair) {
	var buf logdriver.LogEntry

	dec := protoio.NewUint32DelimitedReader(lp.stream, binary.BigEndian, 1e6)
	defer dec.Close()
	defer shutdownLogPair(lp)

	for {
		if !lp.active {
			return
		}

		err := dec.ReadMsg(&buf)
		if err != nil {
			if err == io.EOF {
				return
			} else {
				dec = protoio.NewUint32DelimitedReader(lp.stream, binary.BigEndian, 1e6)
				continue
			}
		}

		err = logMessage(lp, buf.Line)
		if err != nil {
			log.Println("error logging message, dropping it and continuing: " + err.Error())
		}

		buf.Reset()
	}
}
