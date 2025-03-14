package main

import (
	"github.com/paranoiachains/metrics/internal/collector"
)

func main() {
	go collector.UpdateWithInterval(2)
	go collector.SendWithInterval(10)

	select {}
}
