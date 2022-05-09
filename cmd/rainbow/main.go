package main

import (
	"database/sql"
	"embed"
	"flag"
	"fmt"
	"github.com/ATenderholt/rainbow/logging"
	"github.com/ATenderholt/rainbow/settings"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"os"
)

var logger *zap.SugaredLogger

//go:embed migrations/*.sql
var embedMigrations embed.FS

func init() {
	logger = logging.NewLogger()
}

func main() {
	cfg, output, err := settings.FromFlags(os.Args[0], os.Args[1:])
	if err == flag.ErrHelp {
		fmt.Println(output)
		os.Exit(2)
	} else if err != nil {
		fmt.Println("got error:", err)
		fmt.Println("output:\n", output)
		os.Exit(1)
	}

	logger.Info("Starting up ...")

	err = os.MkdirAll(cfg.DataPath(), 0755)
	if err != nil {
		logger.Fatalf("Unable to make data directory: %v", err)
	}

	db, err := cfg.CreateDatabase()
	if err != nil {
		logger.Fatalf("Unable to create database: %v", err)
	}
	defer db.Close()

	initializeDb(db)

	app, err := InjectApp(cfg, db)
	if err != nil {
		logger.Fatalf("Unable to initialize application: %v", err)
	}

	err = app.Start()
	if err != nil {
		logger.Errorf("App failed to start: %v", err)
		app.Shutdown()
	}
}

func initializeDb(db *sql.DB) {
	goose.SetBaseFS(embedMigrations)
	goose.SetLogger(logging.GooseLogger{logger})

	if err := goose.SetDialect("sqlite3"); err != nil {
		panic(err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		panic(err)
	}
}
