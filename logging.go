package loges

import (
	log "github.com/ngmoco/timber"
)

// load given timber config file
func TimberLoadXml(configFile string) {
	defer func() {
		if r := recover(); r != nil {
			println("ERROR in TimberLogging load, proceeding ")
		}
	}()
	log.LoadConfiguration(configFile)
}

// Set Timber logging programattically instead of crappy xml file
func TimberSetLogging(format, logLevel string) {
	var level log.Level
	found := false
	for idx, str := range log.LongLevelStrings {
		if str == logLevel {
			level = log.Level(idx)
			found = true
		}
	}
	if !found {
		println("loglevel not found " + logLevel)
		level = log.Level(6)
	}
	formatter := log.NewPatFormatter(format)
	configLogger := log.ConfigLogger{Level: level, Formatter: formatter}
	configLogger.LogWriter = new(log.ConsoleWriter)
	log.Global.AddLogger(configLogger)
}
