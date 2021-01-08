package master

import (
	"bytes"
	"net/url"
	"os"
	"testing"

	"github.com/centrifugal/centrifuge"
	"go.uber.org/zap"
)

type MemorySink struct {
	*bytes.Buffer
}

func (s *MemorySink) Close() error { return nil }
func (s *MemorySink) Sync() error  { return nil }

type fakeSender struct {
	data []byte
}

func (f *fakeSender) Send(raw centrifuge.Raw) error {
	f.data = raw
	return nil
}

// nolint:gochecknoglobals
var sink *MemorySink

// nolint:gochecknoglobals
var dummyLogger *zap.Logger

func TestMain(m *testing.M) {
	sink = &MemorySink{new(bytes.Buffer)}
	err := zap.RegisterSink("memory", func(*url.URL) (zap.Sink, error) {
		return sink, nil
	})

	if err != nil {
		panic(err)
	}

	logCfg := zap.NewProductionConfig()
	logCfg.OutputPaths = []string{"stdout"}
	dummyLogger, _ = logCfg.Build()

	os.Exit(m.Run())
}
