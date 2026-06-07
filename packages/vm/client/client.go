package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
)

func FirecrakerClient(socketPath string) *http.Client {

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}

	return &http.Client{
		Transport: transport,
	}
}

func CallFirecraker(client *http.Client, method, path string, body []byte) error {
	req, err := http.NewRequest(method, "http://localhost"+path, bytes.NewReader(body))

	if err != nil {
		fmt.Println(err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	fmt.Println("Started the  data")

	fmt.Println("Body ", string(data))

	return nil

}
