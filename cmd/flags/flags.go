package main

import "flag"

var ServerEndpoint string
var ClientEndpoint string
var ReportInterval int
var PollInterval int

func ParseFlags() {
	flag.StringVar(&ServerEndpoint, "a", "127.0.0.1:8080", "Set server endpoint")
	flag.StringVar(&ClientEndpoint, "a", "127.0.0.1:8080", "Set client endpoint")
	flag.IntVar(&ReportInterval, "r", 10, "Set report interval")
	flag.IntVar(&PollInterval, "r", 2, "Set poll interval")

	flag.Parse()
}
