package main

import (
	"github.com/paranoiachains/metrics/internal/collector"
	"github.com/paranoiachains/metrics/internal/flags"
)

func main() {
	flags.ParseAgentFlags()
	flags.ParseEnv()
	if flags.Cfg.PollInterval != 0 {
		flags.PollInterval = flags.Cfg.PollInterval
	}
	if flags.Cfg.ReportInterval != 0 {
		flags.ReportInterval = flags.Cfg.ReportInterval
	}
	if flags.Cfg.Address != "" {
		flags.ClientEndpoint = flags.Cfg.Address
	}
	go collector.UpdateWithInterval(flags.PollInterval)
	go collector.SendWithInterval(flags.ReportInterval, flags.ClientEndpoint)

	select {}
}
