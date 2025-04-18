package flags

import (
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v11"
)

var (
	ServerEndpoint  string
	ClientEndpoint  string
	ReportInterval  int
	PollInterval    int
	EncodingEnabled bool
	StoreInterval   int
	FileStoragePath string
	Restore         bool
	DBEndpoint      string
	ClientKey       string
	ServerKey       string

	Cfg Config

	serverFlags = flag.NewFlagSet("", flag.ExitOnError)
	agentFlags  = flag.NewFlagSet("", flag.ExitOnError)
)

type Config struct {
	DBEndpointEnv   string `env:"DATABASE_DSN"`
	Address         string `env:"ADDRESS"`
	ReportInterval  int    `env:"REPORT_INTERVAL"`
	PollInterval    int    `env:"POLL_INTERVAL"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DBUser          string `env:"DB_USER"`
	DBPassword      string `env:"DB_PASSWORD"`
	DBName          string `env:"DB_NAME"`
	Key             string `env:"KEY"`
}

func ParseEnv() {
	err := env.Parse(&Cfg)
	if err != nil {
		log.Fatal(err)
	}
	if Cfg.PollInterval != 0 {
		PollInterval = Cfg.PollInterval
	}
	if Cfg.ReportInterval != 0 {
		ReportInterval = Cfg.ReportInterval
	}
	if Cfg.StoreInterval != 0 {
		StoreInterval = Cfg.StoreInterval
	}
	if Cfg.FileStoragePath != "" {
		FileStoragePath = Cfg.FileStoragePath
	}
	if !Cfg.Restore {
		Restore = false
	}
	if Cfg.DBEndpointEnv != "" {
		DBEndpoint = Cfg.DBEndpointEnv
	}
	if Cfg.Key != "" {
		ServerKey = Cfg.Key
		ClientKey = Cfg.Key
	}
}

func ParseServerFlags() {
	serverFlags.StringVar(&ServerEndpoint, "a", "localhost:8080", "Set server endpoint")
	serverFlags.IntVar(&StoreInterval, "i", 300, "store interval of metric in seconds")
	serverFlags.StringVar(&FileStoragePath, "f", "tmp/metrics-db.json", "storage file path path")
	serverFlags.BoolVar(&Restore, "r", true, "restore previous metrics")
	serverFlags.StringVar(&DBEndpoint, "d", "", "database endpoint")
	serverFlags.StringVar(&ServerKey, "k", "", "server signature key")
	serverFlags.Parse(os.Args[1:])
}

func ParseAgentFlags() {
	agentFlags.StringVar(&ClientEndpoint, "a", "localhost:8080", "Set client endpoint")
	agentFlags.IntVar(&ReportInterval, "r", 10, "Set report interval")
	agentFlags.IntVar(&PollInterval, "p", 2, "Set poll interval")
	agentFlags.BoolVar(&EncodingEnabled, "e", true, "enable gzip encoding of http requests")
	agentFlags.StringVar(&ClientKey, "k", "", "client signature key")
	agentFlags.Parse(os.Args[1:])
}
