package settings

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const (
	DefaultBasePort   = 8998
	DefaultDbFilename = "db.sqlite3"
	DefaultDataPath   = "data"

	DefaultDockerNetwork  = "rainbow"
	DefaultFunctionsImage = "atenderholt/rainbow-functions:latest"
	DefaultFunctionsName  = "rainbow-functions"
	DefaultFunctionsPort  = 9050

	DefaultMotoImage = "motoserver/moto:3.0.4"
	DefaultMotoName  = "moto"
	DefaultMotoPort  = 8999

	DefaultStorageImage = "atenderholt/rainbow-storage:latest"
	DefaultStorageName  = "rainbow-storage"
	DefaultStoragePort  = 9000

	DefaultSqsImage = "softwaremill/elasticmq:1.3.4"
	DefaultSqsName  = "sqs"
	DefaultSqsPort  = 9324
)

type Container struct {
	Image string
	Name  string
	Port  int
}

type Config struct {
	BasePort   int
	dataPath   string
	dbFileName string
	IsLocal    bool
	Network    string

	Functions Container
	Moto      Container
	Storage   Container
	Sqs       Container
}

func FromFlags(name string, args []string) (*Config, string, error) {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)

	var buf bytes.Buffer
	flags.SetOutput(&buf)

	var cfg Config
	flags.IntVar(&cfg.BasePort, "base-port", DefaultBasePort, "Port to start HTTP")
	flags.StringVar(&cfg.dataPath, "data-path", DefaultDataPath, "Path to persist data")
	flags.StringVar(&cfg.dbFileName, "db", DefaultDbFilename, "Database file for persisting configuration")
	flags.BoolVar(&cfg.IsLocal, "local", true, "Application is running locally (vs. in container)")

	flags.StringVar(&cfg.Network, "network", DefaultDockerNetwork, "Network to run docker containers")

	flags.StringVar(&cfg.Functions.Image, "functions-image", DefaultFunctionsImage, "Docker image for functions")
	flags.StringVar(&cfg.Functions.Name, "functions-name", DefaultFunctionsName, "Docker container name for functions")
	flags.IntVar(&cfg.Functions.Port, "functions-port", DefaultFunctionsPort, "Port for functions running on localhost")

	flags.StringVar(&cfg.Moto.Image, "moto-image", DefaultMotoImage, "Docker image for moto")
	flags.StringVar(&cfg.Moto.Name, "moto-name", DefaultMotoName, "Docker container name for moto")
	flags.IntVar(&cfg.Moto.Port, "moto-port", DefaultMotoPort, "Port for moto running on localhost")

	flags.StringVar(&cfg.Storage.Image, "storage-image", DefaultStorageImage, "Docker image for storage")
	flags.StringVar(&cfg.Storage.Name, "storage-name", DefaultStorageName, "Docker container name for storage")
	flags.IntVar(&cfg.Storage.Port, "storage-port", DefaultStoragePort, "Port for storage running on localhost")

	flags.StringVar(&cfg.Sqs.Image, "sqs-image", DefaultSqsImage, "Docker image for sqs")
	flags.StringVar(&cfg.Sqs.Name, "sqs-name", DefaultSqsName, "Docker container name for sqs")
	flags.IntVar(&cfg.Sqs.Port, "sqs-port", DefaultSqsPort, "Port for sqs running on localhost")

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

func (config Config) FunctionsHost() string {
	if config.IsLocal {
		return "localhost:" + strconv.Itoa(config.Functions.Port)
	}

	return config.Functions.Name + ":" + strconv.Itoa(9050)
}

func (config Config) MotoHost() string {
	if config.IsLocal {
		return "localhost:" + strconv.Itoa(config.Moto.Port)
	}

	return config.Moto.Name + ":" + strconv.Itoa(5000)
}

func (config Config) StorageHost() string {
	if config.IsLocal {
		return "localhost:" + strconv.Itoa(config.Storage.Port)
	}

	return config.Storage.Name + ":" + strconv.Itoa(9000)
}

func (config Config) SqsHost() string {
	if config.IsLocal {
		return "localhost:" + strconv.Itoa(config.Sqs.Port)
	}

	return config.Sqs.Name + ":" + strconv.Itoa(9324)
}
