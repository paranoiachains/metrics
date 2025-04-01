package main

import (
	"github.com/paranoiachains/metrics/internal/collector"
	"github.com/paranoiachains/metrics/internal/flags"
)

func main() {
	flags.ParseAgentFlags()
	if flags.Cfg.Address != "" {
		flags.ClientEndpoint = flags.Cfg.Address
	}
	collector.ClearMetrics()
	go collector.UpdateWithInterval(flags.PollInterval)
	go collector.SendWithInterval(flags.ReportInterval, flags.ClientEndpoint)

	select {}
}
