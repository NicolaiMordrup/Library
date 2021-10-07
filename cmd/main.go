package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	library "github.com/NicolaiMordrup/library"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"
)

func main() {
	// Configuration
	connstr := "file:librarystorage.db"
	if envVal := os.Getenv("SQLITE_DB_CONN"); envVal != "" {
		connstr = envVal
	}
	portStr := "8000"
	if envVal := os.Getenv("SERVER_PORT"); envVal != "" {
		portStr = envVal
	}
	minDurationBetweenUpdatesStr := "10s"
	if envVal := os.Getenv("MIN_DURATION_BETWEEN_UPDATES"); envVal != "" {
		minDurationBetweenUpdatesStr = envVal
	}
	minDurationBetweenUpdates, err := time.ParseDuration(minDurationBetweenUpdatesStr)
	check(err, "failed to parse min duration between updates")
	_ = minDurationBetweenUpdates

	// Setup logger
	structuredLogger, _ := zap.NewProduction()
	log := structuredLogger.Sugar()

	// Connect to database
	// Note(sn): add storage constructor stuff here
	// Note(sn): add logger to database (call it log)
	db, err := library.NewDB(connstr)
	check(err, "failed to open sqlite connection")
	check(library.EnsureSchema(db), "migration failed")

	// Initialize and start server
	// Note(sn): add min duration to server constructor
	// Note(sn): add logger to server
	myServer := library.NewServer(db)
	addr := fmt.Sprintf(":%v", portStr)
	log.Infow("starting server",
		"addr", addr,
	)
	log.Fatal(http.ListenAndServe(addr, myServer))
}

func check(err error, msg string) {
	if err != nil {
		fmt.Printf("%v, err: %v\n", msg, err)
		os.Exit(1)
	}
}
