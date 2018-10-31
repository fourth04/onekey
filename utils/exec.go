package utils

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Executor execute input string
func Execute(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("you need to pass the something arguments")
	} else if s == "quit" || s == "exit" {
		log.Println("Bye!")
		os.Exit(0)
	}

	cmd := exec.Command("bash", "-c", s)
	// log.Println("Command：", "bash", "-c", s)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func ExecuteAndPrintImmediately(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("you need to pass the something arguments")
	} else if s == "quit" || s == "exit" {
		log.Println("Bye!")
		os.Exit(0)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd := exec.Command("bash", "-c", s)
	// log.Println("Command：", "bash", "-c", s)
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	var errStdout, errStderr error
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)
	err := cmd.Start()
	if err != nil {
		return err
	}
	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
	}()
	go func() {
		_, errStderr = io.Copy(stderr, stderrIn)
	}()
	err = cmd.Wait()
	if err != nil {
		return err
	}
	if errStdout != nil {
		return errStdout
	}
	if errStderr != nil {
		return errStderr
	}
	return nil
}

// ExecuteAndGetResult execute input string and return the echo
func ExecuteAndGetResult(s string) (string, string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", "", errors.New("you need to pass the something arguments")
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", s)
	// log.Println("Command：", "bash", "-c", s)
	cmd.Stdin = os.Stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", "", err
	}
	// log.Println("Execute Return：", strings.TrimSpace(string(stdout.Bytes())))
	return strings.TrimSpace(string(stdout.Bytes())), strings.TrimSpace(string(stderr.Bytes())), nil
}

// ExecuteAndGetResult execute input string and return the echo
func ExecuteAndGetResultCombineError(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("you need to pass the something arguments")
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", s)
	log.Println("Execute Command：", "bash", "-c", s)
	cmd.Stdin = os.Stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	stderrStr := strings.TrimSpace(string(stderr.Bytes()))
	if stderrStr != "" {
		return "", errors.New(stderrStr)
	}
	if err != nil {
		return "", err
	}
	log.Println("Execute Return：", strings.TrimSpace(string(stdout.Bytes())))
	return strings.TrimSpace(string(stdout.Bytes())), nil
}
