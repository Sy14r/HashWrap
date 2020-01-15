package main

// https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html
// to run:
// go run 03-live-progress-and-capture-v3.go

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

func main() {
	intervalTime, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Invalid interval provided")
	}
	cmd := exec.Command(os.Args[2], os.Args[3:]...)

	// var stdoutBuf, stderrBuf bytes.Buffer
	var stdoutBuf bytes.Buffer
	stdoutIn, _ := cmd.StdoutPipe()
	//stderrIn, _ := cmd.StderrPipe()
	stdinIn, _ := cmd.StdinPipe()

	// var errStdout error
	stdout := io.MultiWriter(&stdoutBuf)
	// stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	//stderr := io.MultiWriter(os.Stderr, &stderrBuf)
	err = cmd.Start()
	if err != nil {
		log.Fatalf("cmd.Start() failed with '%s'\n", err)
	}

	var wg sync.WaitGroup

	wg.Add(1)

	// Go func to get stdout passed through and finish the WaitGroup when exe is done
	go func() {
		// _, errStdout = io.Copy(stdout, stdoutIn)
		io.Copy(stdout, stdoutIn)
		wg.Done()
	}()

	// Go func to repeatedly send the status check keypress and write output to disk (for S3 upload)
	go func() {
		defer stdinIn.Close()
		sum := 1
		for sum < 2 {
			// in loop will want to check for tasking from s3 (can manage this with an external script and just have this look for the local file of hashcat.externalsignal and process that as needed?) - this is a future development
			if _, err := os.Stat("./hashwrap.pause"); err == nil {
				io.WriteString(stdinIn, "c")
			} else if os.IsNotExist(err) {
				io.WriteString(stdinIn, "s")
			} else {
				io.WriteString(stdinIn, "s")
			}
			//data := strings.Split(stdoutBuf.String(), "\n")
			d1 := stdoutBuf.Bytes()
			ioutil.WriteFile("./hashcat.status", d1, 0644)
			//fmt.Println()
			stdoutBuf.Reset()
			time.Sleep(time.Duration(intervalTime) * time.Second)
		}
	}()

	//_, errStderr = io.Copy(stderr, stderrIn)
	wg.Wait()

	err = cmd.Wait()
	// if err != nil {
	// 	log.Fatalf("cmd.Run() failed with %s\n", err)
	// }
	// if errStdout != nil || errStderr != nil {
	// 	log.Fatal("failed to capture stdout or stderr\n")
	// }

	// outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	// outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	// fmt.Printf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)
	ioutil.WriteFile("./hashcat.status", stdoutBuf.Bytes(), 0644)
}
