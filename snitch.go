package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {
	var args []string
	argsLen := len(os.Args)
	if argsLen < 2 {
		log.Fatal("Not enough arguments.")
	} else if argsLen > 2 {
		args = os.Args[2:]
	}

	cmd := exec.Command(os.Args[1], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatalf("Couldn't start command : %s", err)
	}

	childExited := make(chan error, 1)
	go func() {
		childExited <- cmd.Wait()
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig)
	for {
		select {
		case signal := <-sig:
			// ignore SIGCHLD
			if signal == syscall.SIGCHLD {
				continue
			}

			syscall.Kill(cmd.Process.Pid, signal.(syscall.Signal))

		case err := <-childExited:
			if err == nil {
				os.Exit(0)
			}

			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					os.Exit(status.ExitStatus())
				}
			}

			log.Fatalf("Command didn't run as expected : %s", err)
		}
	}
}
