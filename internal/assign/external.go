package assign

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

func RunExternalAssigner(input InputFormat, hallRequestAssignerPath string) OutputFormat {
	encodedInput, encodingError := json.Marshal(input)
	if encodingError != nil {
		panic(fmt.Sprintf("Failed to encode input for assigner: %v", encodingError))
	}

	command := exec.Command(hallRequestAssignerPath, "--input", string(encodedInput))

	encodedOutput, assignmentError := command.Output()
	if assignmentError != nil {
		panic(fmt.Sprintf("Failed to run assigner binary: %v", assignmentError))
	}

	var output OutputFormat
	if decodingError := json.Unmarshal(encodedOutput, &output); decodingError != nil {
		panic(fmt.Sprintf("Failed to decode assigner output: %v", decodingError))
	}

	return output
}
