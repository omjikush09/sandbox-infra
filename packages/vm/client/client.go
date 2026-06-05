package client

import (
	"context"
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

func CallFirecraker(){
	
}


