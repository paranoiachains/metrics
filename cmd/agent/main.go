package main

import "github.com/paranoiachains/metrics/internal/sender"

func main() {
	sender.MyMetrics.Clear()
	go sender.UpdateWithInterval(2)
	go sender.SendMetricsWithInterval(10)

	select {}
}
