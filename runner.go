package katsubushi

import (
	"io"
	"log"
	"time"

	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
)

type Runner struct {
	Interval time.Duration

	generator   IdGenerator
	replacement string
	cmd         []string
}

var TrapSignals = []os.Signal{
	syscall.SIGHUP,
	syscall.SIGINT,
	syscall.SIGTERM,
	syscall.SIGQUIT}

func NewRunner(generator IdGenerator, replacement string, cmd []string) *Runner {
	return &Runner{
		Interval:    time.Minute,
		generator:   generator,
		replacement: replacement,
		cmd:         cmd,
	}
}

func (r *Runner) Run() error {
	idInfo, err := r.generator.New()
	if err != nil {
		return err
	}

	cmd, err := r.startCommand(idInfo.Id)
	if err != nil {
		return err
	}
	cmdCh := make(chan error)
	go func() {
		cmdCh <- cmd.Wait()
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, TrapSignals...)

	for {
		select {
		case <-time.After(r.Interval):
			// extend expire time
			if newIdInfo, err := r.generator.Update(idInfo.Id); err == nil {
				idInfo = newIdInfo
				log.Print("extend expire time to %s", idInfo.ExpireAt)
			} else {
				log.Print(err)
			}

			now := time.Now()
			if idInfo.ExpireAt.Before(now.Add(r.Interval)) {
				// id is expired. terminate child process...
				log.Print("expired")
				cmd.Process.Signal(syscall.SIGTERM)
				idInfo.ExpireAt = now.Add(-1 * time.Nanosecond)
			}
		case s := <-signalCh:
			cmd.Process.Signal(s) // forward to child
		case cmdErr := <-cmdCh:
			if time.Now().Before(idInfo.ExpireAt) {
				r.generator.Delete(idInfo.Id)
			}
			return cmdErr
		}
	}

	return nil
}

func (r *Runner) startCommand(id int) (*exec.Cmd, error) {
	args := make([]string, len(r.cmd)-1)
	for i, arg := range r.cmd[1:] {
		if arg == r.replacement {
			args[i] = strconv.Itoa(id)
		} else {
			args[i] = arg
		}
	}

	cmd := exec.Command(r.cmd[0], args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	defer stdin.Close()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	go io.Copy(stdin, os.Stdin)
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	cmd.Start()
	return cmd, nil
}
