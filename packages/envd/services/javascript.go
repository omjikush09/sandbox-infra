package services

import (
	"os"
	"os/exec"

	"github.com/omjikush09/sandboxing-infra/packages/envd/models"
)

var filePath = "./code.js"

func WriteToFile(code *models.Execute) error {
	return os.WriteFile(filePath, []byte(code.Code), 0644)
}

func ExecuteJS() (string, error) {

	bunCommand := exec.Command("bun", "run", filePath)

	out, err := bunCommand.CombinedOutput()

	return string(out), err
}
