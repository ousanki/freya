package log

import (
	go_logger "github.com/phachon/go-logger"
)

var logger *go_logger.Logger

func init() {
	logger = go_logger.NewLogger()
	_ = logger.Detach("console")
	fileConfig := &go_logger.FileConfig{
		Filename: "./logfile/access.log",
		LevelFileName: map[int]string{
			logger.LoggerLevel("error"):   "./logfile/error.log",
			logger.LoggerLevel("warning"): "./logfile/warn.log",
			logger.LoggerLevel("info"):    "./logfile/info.log",
			logger.LoggerLevel("debug"):   "./logfile/debug.log",
		},
		MaxSize:    1024 * 1024,
		MaxLine:    20000,
		DateSlice:  "d",
		JsonFormat: false,
		Format:     "%timestamp_format% [%level_string%](%file%:%line%): %body%",
	}
	_ = logger.Attach("file", go_logger.LOGGER_LEVEL_DEBUG, fileConfig)

	consoleConfig := &go_logger.ConsoleConfig{
		Color:      false,
		JsonFormat: false,
		Format:     "%millisecond_format% [%level_string%] [%file%:%line%] %body%",
	}

	_ = logger.Attach("console", go_logger.LOGGER_LEVEL_DEBUG, consoleConfig)
}

func GetLogger() *go_logger.Logger {
	return logger
}
