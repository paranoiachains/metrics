package flags

import (
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v11"
)

var ServerEndpoint string
var ClientEndpoint string
var ReportInterval int
var PollInterval int
var EncodingEnabled bool
var StoreInterval int
var FileStoragePath string
var Restore bool

var Cfg Config

type Config struct {
	Address         string `env:"ADDRESS"`
	ReportInterval  int    `env:"REPORT_INTERVAL"`
	PollInterval    int    `env:"POLL_INTERVAL"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

func init() {
	ParseEnv()
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
}

func ParseServerFlags() {
	serverFlags := flag.NewFlagSet("", flag.ExitOnError)
	serverFlags.StringVar(&ServerEndpoint, "a", "localhost:8080", "Set server endpoint")
	serverFlags.IntVar(&StoreInterval, "i", 300, "store interval of metric in seconds")
	serverFlags.StringVar(&FileStoragePath, "f", "/tmp/metrics-db.json", "storage file path path")
	serverFlags.BoolVar(&Restore, "r", true, "restore previous metrics")
	serverFlags.Parse(os.Args[1:])
	FileStoragePath = "../.." + FileStoragePath
}

func ParseAgentFlags() {
	agentFlags := flag.NewFlagSet("", flag.ExitOnError)
	agentFlags.StringVar(&ClientEndpoint, "a", "localhost:8080", "Set client endpoint")
	agentFlags.IntVar(&ReportInterval, "r", 10, "Set report interval")
	agentFlags.IntVar(&PollInterval, "p", 2, "Set poll interval")
	agentFlags.BoolVar(&EncodingEnabled, "e", true, "enable gzip encoding of http requests")
	agentFlags.Parse(os.Args[1:])
}
