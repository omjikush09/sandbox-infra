package main

import (
	"fmt"
	"net/http"

	"github.com/omjikush09/sandboxing-infra/packages/vm/client"
)

func main() {

	path := "/tmp/firecracker.socket"

	httpClient := client.FirecrakerClient(path)

	//Get machine config
	// body:=[]byte(``)
	err := client.CallFirecraker(httpClient, http.MethodGet, "/", nil)
	if err != nil {
		fmt.Println(err)
	}
	

}
