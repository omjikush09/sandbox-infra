package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type std struct {
	Kind string
	data []byte
}

func ExecuteCmd(command []string) (string, string, error) {
	cmd := exec.Command(command[0], command[1:]...)
	var stdOut bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &errOut

	err := cmd.Run()

	if err != nil {
		return "", "", fmt.Errorf("Failed able to run the command")
	}
	return stdOut.String(), errOut.String(), nil
}

func ExecuteCmdBackground(command []string, c context.Context) (chan []byte, error) {
	cmd := exec.Command(command[0], command[1:]...)
	stdOut, err := cmd.StdoutPipe()
	stdErr, err := cmd.StderrPipe()
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("Failed to start cmd")
	}

	event := make(chan []byte)

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				buf := make([]byte, 2024)
				n, err := stdOut.Read(buf)
				if err != nil {
					continue
				}
				if n > 0 {
					data := std{data: buf[:n], Kind: "STDOUT"}

					byt, err := json.Marshal(data)
					if err != nil {
						event <- byt
					}
				}

				bufe := make([]byte, 2024)

				n, err = stdErr.Read(bufe)
				if err != nil {
					continue
				}
				if n > 0 {
					data := std{data: buf[:n], Kind: "STERR"}
					byt, err := json.Marshal(data)
					if err != nil {
						event <- byt
					}
				}

			case <-c.Done():
				return
			}
		}
	}()

	return event, nil
}
