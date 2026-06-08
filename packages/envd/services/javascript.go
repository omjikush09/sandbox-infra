package services

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/omjikush09/sandboxing-infra/packages/envd/models"
)

func ExecuteJS(code *models.Execute) (string, error) {
	filePath := "./code.js"
	_, err := os.Create(filePath)
	if err != nil {
		fmt.Println(err)
	}
	err = os.WriteFile("./code.js", []byte(code.Code), 0644)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile("./code.js")
	if err != nil {
		panic(err)
	} 
	fmt.Println(string(data))
	bunCommand := exec.Command("bun", "run", filePath)

	out, err := bunCommand.CombinedOutput()

	return string(out), err
	//

}
