package main

import (
	"github.com/paranoiachains/metrics/internal/collector"
	"github.com/paranoiachains/metrics/internal/flags"
)

func main() {
	flags.ParseAgentFlags()
	flags.ParseEnv()
	if flags.Cfg.POLL_INTERVAL != 0 {
		flags.PollInterval = flags.Cfg.POLL_INTERVAL
	}
	if flags.Cfg.REPORT_INTERVAL != 0 {
		flags.ReportInterval = flags.Cfg.REPORT_INTERVAL
	}
	if flags.Cfg.ADDRESS != "" {
		flags.ClientEndpoint = flags.Cfg.ADDRESS
	}
	go collector.UpdateWithInterval(flags.PollInterval)
	go collector.SendWithInterval(flags.ReportInterval, flags.ClientEndpoint)

	select {}
}
