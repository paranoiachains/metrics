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
var Cfg Config

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func ParseEnv() {
	err := env.Parse(&Cfg)
	if err != nil {
		log.Fatal(err)
	}
}

func ParseServerFlags() {
	serverFlags := flag.NewFlagSet("", flag.ExitOnError)
	serverFlags.StringVar(&ServerEndpoint, "a", "localhost:8080", "Set server endpoint")
	serverFlags.Parse(os.Args[1:])
}

func ParseAgentFlags() {
	clientFlags := flag.NewFlagSet("", flag.ExitOnError)
	clientFlags.StringVar(&ClientEndpoint, "a", "localhost:8080", "Set client endpoint")
	clientFlags.IntVar(&ReportInterval, "r", 10, "Set report interval")
	clientFlags.IntVar(&PollInterval, "p", 2, "Set poll interval")
	clientFlags.Parse(os.Args[1:])
}
