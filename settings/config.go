package settings

import (
	"bytes"
	"database/sql"
	"flag"
	"log"
	"os"
	"path/filepath"
)

const (
	DefaultDbFilename = "db.sqlite3"
	DefaultDataPath   = "data"
)

type Config struct {
	dataPath   string
	dbFileName string
}

func FromFlags(name string, args []string) (*Config, string, error) {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)

	var buf bytes.Buffer
	flags.SetOutput(&buf)

	var cfg Config
	flags.StringVar(&cfg.dataPath, "data-path", DefaultDataPath, "Path to persist data")
	flags.StringVar(&cfg.dbFileName, "db", DefaultDbFilename, "Database file for persisting configuration")

	err := flags.Parse(args)
	if err != nil {
		return nil, buf.String(), err
	}

	return &cfg, buf.String(), err
}

func (config *Config) CreateDatabase() *sql.DB {
	connStr := "file:" + filepath.Join(config.dataPath)
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		log.Panicf("unable to open database %s: %v", connStr, err)
	}

	err = db.Ping()
	if err != nil {
		log.Panicf("unable to ping database %s: %v", connStr, err)
	}

	return db
}

func (config *Config) DataPath() string {
	if config.dataPath[0] == '/' {
		return config.dataPath
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return filepath.Join(cwd, config.dataPath)
}
