// pkg/utils/logger.go
// **************************************************************
// Author: Tyler Laudenslager & Tyler Nazzaro
// Purpose: Custom logger for enhanced logging capabilities.
// **************************************************************

package utils

import (
	"io"
	"log"
	"os"
)

// InitLogger initializes and returns a custom logger.
// It writes logs to both standard output and a log file.
func InitLogger() *log.Logger {
	logFile, err := os.OpenFile("rps_game.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v\n", err)
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	customLogger := log.New(multiWriter, "RPS_LOG: ", log.Ldate|log.Ltime|log.Lshortfile)
	return customLogger
}
