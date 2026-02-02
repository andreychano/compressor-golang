package logger

import gotoolslog "github.com/shanth1/gotools/log"

var Log gotoolslog.Logger

func Init(cfg gotoolslog.Config) {
	Log = gotoolslog.NewFromConfig(cfg)
}
