package main

// https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html
// to run:
// go run 03-live-progress-and-capture-v3.go

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

func main() {
	cmd := exec.Command(os.Args[1], os.Args[2:]...)

	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	stdinIn, _ := cmd.StdinPipe()

	var errStdout, errStderr error
	stdout := io.MultiWriter(&stdoutBuf)
	// stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)
	err := cmd.Start()
	if err != nil {
		log.Fatalf("cmd.Start() failed with '%s'\n", err)
	}

	var wg sync.WaitGroup

	wg.Add(1)

	// Go func to get stdout passed through and finish the WaitGroup when exe is done
	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
		wg.Done()
	}()

	// Go func to repeatedly send the status check keypress and write output to disk (for S3 upload)
	go func() {
		defer stdinIn.Close()
		sum := 1
		for sum < 2 {
			stdoutBuf.Reset()
			time.Sleep(5 * time.Second)
			io.WriteString(stdinIn, "s")
			//data := strings.Split(stdoutBuf.String(), "\n")
			d1 := stdoutBuf.Bytes()
			ioutil.WriteFile("./hashcat.status", d1, 0644)
			//fmt.Println()
		}
	}()

	_, errStderr = io.Copy(stderr, stderrIn)
	wg.Wait()

	err = cmd.Wait()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	if errStdout != nil || errStderr != nil {
		log.Fatal("failed to capture stdout or stderr\n")
	}

	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	fmt.Printf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)
}
