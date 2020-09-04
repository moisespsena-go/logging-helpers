package logging_helpers

import (
	"strings"

	logrotate "github.com/moisespsena-go/glogrotation"
	"github.com/moisespsena-go/logging/backends"
)

func Rotates(backend *backends.FileBackend, options ...logrotate.Options) {
	if !strings.Contains(backend.WriteCloserBackend.Name, "@rotation") {
		rotator := logrotate.New(backend.Path(), options...)
		old := backend.WriteCloserBackend
		old.Close()
		backend.WriteCloserBackend = backends.NewWriteCloserBackend("file@rotation:"+rotator.Path, rotator, old.Async)
	}
}
