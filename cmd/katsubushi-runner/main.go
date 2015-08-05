package main

import (
	"flag"

	katsubushi "github.com/shogo82148/go-katsubushi-katshubushi"
)

func main() {
	var server string
	var replacement string
	flag.StringVar(&server, "server", "", "worker id server")
	flag.StringVar(&replacement, "replacement", "worker-id", "replacement text for worker id")
	flag.Parse()

	generator := katsubushi.NewClient(server)
	runner := katsubushi.NewRunner(generator, replacement, flag.Args())
	runner.Run()
}
