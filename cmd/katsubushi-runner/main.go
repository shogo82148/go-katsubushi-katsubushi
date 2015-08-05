package main

import (
	"flag"
	"time"

	katsubushi "github.com/shogo82148/go-katsubushi-katshubushi"
)

func main() {
	var server string
	var replacement string
	var interval time.Duration
	flag.StringVar(&server, "server", "", "worker id server")
	flag.StringVar(&replacement, "replacement", "worker-id", "replacement text for worker id")
	flag.DurationVar(&interval, "interval", time.Minute, "interval duration time")
	flag.Parse()

	generator := katsubushi.NewClient(server)
	runner := katsubushi.NewRunner(generator, replacement, flag.Args())
	runner.Interval = interval
	runner.Run()
}
