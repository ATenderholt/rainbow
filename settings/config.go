package settings

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultBasePort   = 8998
	DefaultDbFilename = "db.sqlite3"
	DefaultDataPath   = "data"
)

type Config struct {
	BasePort   int
	dataPath   string
	dbFileName string
}

func FromFlags(name string, args []string) (*Config, string, error) {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)

	var buf bytes.Buffer
	flags.SetOutput(&buf)

	var cfg Config
	flags.IntVar(&cfg.BasePort, "base-port", DefaultBasePort, "Port to start HTTP")
	flags.StringVar(&cfg.dataPath, "data-path", DefaultDataPath, "Path to persist data")
	flags.StringVar(&cfg.dbFileName, "db", DefaultDbFilename, "Database file for persisting configuration")

	err := flags.Parse(args)
	if err != nil {
		return nil, buf.String(), err
	}

	return &cfg, buf.String(), err
}

func (config Config) DatabaseConnection() string {
	return filepath.Join(config.dataPath, config.dbFileName)
}

func (config *Config) CreateDatabase() (*sql.DB, error) {
	connStr := "file:" + filepath.Join(config.dataPath, config.dbFileName)
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		e := fmt.Errorf("unable to open database %s: %v", connStr, err)
		return nil, e
	}

	err = db.Ping()
	if err != nil {
		e := fmt.Errorf("unable to ping database %s: %v", connStr, err)
		return nil, e
	}

	return db, nil
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
