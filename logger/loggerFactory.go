package logger

import (
	"fmt"
)

type Factory struct {
	base Config
}

func NewFactory(base Config) *Factory {
	return &Factory{base: base}
}

func (f *Factory) New(logFileName string) (*Logger, error) {
	cfg := f.base
	cfg.Output = &LogFile{
		LogDir:      f.base.Output.LogDir,
		LogFileName: fmt.Sprintf("%s.log", logFileName),
	}
	return NewLogger(cfg)
}

func (f *Factory) NewWith(logFileName string, opts func(*Config)) (*Logger, error) {
	cfg := f.base
	opts(&cfg)
	cfg.Output = &LogFile{
		LogDir:      f.base.Output.LogDir,
		LogFileName: fmt.Sprintf("%s.log", logFileName),
	}
	return NewLogger(cfg)
}
