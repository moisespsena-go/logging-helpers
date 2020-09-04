package logging_helpers

import (
	"net/url"
	"strings"

	logrotate "github.com/moisespsena-go/glogrotation"

	"github.com/apex/log"

	"github.com/mitchellh/mapstructure"

	"github.com/moisespsena-go/logging"
	"github.com/moisespsena-go/logging/backends"
)

type LogLevel struct {
	Level string
}

var levels = map[string]logging.Level{
	"CRITICAL": logging.CRITICAL,
	"C":        logging.CRITICAL,
	"ERROR":    logging.ERROR,
	"E":        logging.ERROR,
	"WARNING":  logging.WARNING,
	"W":        logging.WARNING,
	"NOTICE":   logging.NOTICE,
	"N":        logging.NOTICE,
	"INFO":     logging.INFO,
	"I":        logging.INFO,
	"DEBUG":    logging.DEBUG,
	"D":        logging.DEBUG,
}

func (ll LogLevel) GetLevel(defaul ...logging.Level) logging.Level {
	if l, ok := levels[strings.ToUpper(ll.Level)]; ok {
		return l
	}
	for _, d := range defaul {
		return d
	}
	return logging.DEBUG
}

type ModuleLoggingBackendConfig struct {
	LogLevel `yaml:",inline"`
	Dst      string
	Options  map[string]interface{}
}

type ModuleLoggingConfig struct {
	LogLevel    `yaml:",inline"`
	Name        string
	Backends    []ModuleLoggingBackendConfig `yaml:"backends"`
	ErrBackends []ModuleLoggingBackendConfig `yaml:"err_backends" mapstructure:"err_backends"`
	Options     map[string]interface{}
}

func (this ModuleLoggingConfig) backendFor(items []ModuleLoggingBackendConfig) (results []logging.BackendCloser) {
	for i, b := range items {
		if strings.HasPrefix(b.Dst, "http:") || strings.HasPrefix(b.Dst, "https:") {
			var opts backends.HttpOptions
			opts.Async = true
			err := mapstructure.Decode(b.Options, &opts)
			if err != nil {
				log.Errorf("parse http options for backend #%d `%s` failed: %s", i, b.Dst, err.Error())
				continue
			}
			URL, err := url.Parse(b.Dst)
			if err != nil {
				log.Errorf("parse url for backend #%d `%s` failed: %s", i, b.Dst, err.Error())
				continue
			}
			bce := backends.NewHttpBackend(*URL, opts, nil)
			results = append(results, bce)
		} else if b.Dst == "-" || b.Dst == "_" {
			results = append(results, logging.NewBackendClose(logging.DefaultBackendProxy()))
		} else {
			var opts backends.FileOptions
			opts.Async = true
			err := mapstructure.Decode(b.Options, &opts)
			if err != nil {
				log.Errorf("parse file options for backend #%d `%s` failed: %s", i, b.Dst, err.Error())
				continue
			}
			bce, err := backends.NewFileBackend(b.Dst, opts)
			if err != nil {
				log.Errorf("create file backend for backend #%d `%s` failed: %s", i, b.Dst, err.Error())
				continue
			}
			if disabled, ok := b.Options["rotate_disabled"]; !ok || !disabled.(bool) {
				var (
					cfg  logrotate.Config
					opts logrotate.Options
				)
				if rotateConfig, ok := b.Options["rotate"]; ok {
					if err := mapstructure.Decode(rotateConfig, &cfg); err != nil {
						log.Errorf("parse file rotate options for backend #%d `%s` failed: %s", i, b.Dst, err.Error())
					} else if opts, err = cfg.Options(); err != nil {
						log.Errorf("bad file rotate options for backend #%d `%s` failed: %s", i, b.Dst, err.Error())
					}
				}
				Rotates(bce, opts)
			}
			results = append(results, bce)
		}
	}
	return
}

func (this ModuleLoggingConfig) Backend() (results []logging.BackendCloser) {
	return this.backendFor(this.Backends)
}

func (this ModuleLoggingConfig) ErrBackend() (results []logging.BackendCloser) {
	return this.backendFor(this.ErrBackends)
}

func (this ModuleLoggingConfig) backendPrinterFor(items []ModuleLoggingBackendConfig) (results []logging.BackendPrintCloser) {
	for i, b := range items {
		if strings.HasPrefix(b.Dst, "http:") || strings.HasPrefix(b.Dst, "https:") {
			var opts backends.HttpOptions
			opts.Async = true
			err := mapstructure.Decode(b.Options, &opts)
			if err != nil {
				log.Errorf("parse http options for backend #%d `%s` failed: %s", i, b.Dst, err.Error())
				continue
			}
			URL, err := url.Parse(b.Dst)
			if err != nil {
				log.Errorf("parse url for backend #%d `%s` failed: %s", i, b.Dst, err.Error())
				continue
			}
			bce := backends.NewHttpBackend(*URL, opts, nil)
			results = append(results, bce)
		} else {
			var opts backends.FileOptions
			opts.Async = true
			err := mapstructure.Decode(b.Options, &opts)
			if err != nil {
				log.Errorf("parse file options for backend #%d `%s` failed: %s", i, b.Dst, err.Error())
				continue
			}
			bce, err := backends.NewFileBackend(b.Dst, opts)
			if err != nil {
				log.Errorf("create file backend for backend #%d `%s` failed: %s", i, b.Dst, err.Error())
				continue
			}
			if disabled, ok := b.Options["rotate_disabled"]; !ok || !disabled.(bool) {
				var (
					cfg  logrotate.Config
					opts logrotate.Options
				)
				if rotateConfig, ok := b.Options["rotate"]; ok {
					if err := mapstructure.Decode(rotateConfig, &cfg); err != nil {
						log.Errorf("parse file rotate options for backend #%d `%s` failed: %s", i, b.Dst, err.Error())
					} else if opts, err = cfg.Options(); err != nil {
						log.Errorf("bad file rotate options for backend #%d `%s` failed: %s", i, b.Dst, err.Error())
					}
				}
				Rotates(bce, opts)
			}
			results = append(results, bce)
		}
	}
	return
}

func (this ModuleLoggingConfig) BackendPrinter() (results []logging.BackendPrintCloser) {
	return this.backendPrinterFor(this.Backends)
}

func (this ModuleLoggingConfig) ErrBackendPrinter() (results []logging.BackendPrintCloser) {
	return this.backendPrinterFor(this.ErrBackends)
}

type LoggingConfig struct {
	LogLevel `yaml:",inline"`
	Modules  []ModuleLoggingConfig
}
