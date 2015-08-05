package katsubushi

import (
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"gopkg.in/Sirupsen/logrus.v0"
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
	err := r.run()
	if err != nil {
		logrus.Error(err)
	}
	return err
}

func (r *Runner) run() error {
	idInfo, err := r.generator.New()
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"id":          idInfo.Id,
		"released_at": idInfo.ReleasedAt,
		"expire_at":   idInfo.ExpireAt,
	}).Info("post")
	defer func() {
		if time.Now().Before(idInfo.ExpireAt) {
			r.generator.Delete(idInfo.Id)
			logrus.WithField("id", idInfo.Id).Info("delete")
		}
	}()

	args := make([]string, len(r.cmd)-1)
	for i, arg := range r.cmd[1:] {
		if arg == r.replacement {
			args[i] = strconv.Itoa(idInfo.Id)
		} else {
			args[i] = arg
		}
	}

	cmd := exec.Command(r.cmd[0], args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer stdin.Close()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	go io.Copy(stdin, os.Stdin)
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	cmd.Start()
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
				logrus.WithFields(logrus.Fields{
					"id":          idInfo.Id,
					"released_at": idInfo.ReleasedAt,
					"expire_at":   idInfo.ExpireAt,
				}).Info("put")
			} else {
				logrus.Error(err)
			}

			now := time.Now()
			if idInfo.ExpireAt.Before(now.Add(r.Interval)) {
				// id is expired. terminate child process...
				logrus.WithField("id", idInfo.Id).Info("expired")
				cmd.Process.Signal(syscall.SIGTERM)
				idInfo.ExpireAt = now.Add(-1 * time.Nanosecond)
			}
		case s := <-signalCh:
			cmd.Process.Signal(s) // forward to child
			logrus.WithField("signal", s).Info("signal")
		case cmdErr := <-cmdCh:
			logrus.WithField("status", cmdErr).Info("exit")
			return cmdErr
		}
	}

	return nil
}
