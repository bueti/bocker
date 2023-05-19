// Package logger contains functionallity to log commands run by the TUI to a log file.
// Taken (and adapted) from https://github.com/zackproser/bubbletea-stages
package logger

import (
	"fmt"
	"os"
	"time"
)

// commandLog is rendered when the deployment encounters an error. It retains a log of all the "commands" that were run in the course of deploying the example
// "commands" are intentionally in air-quotes here because this also includes things like checking for the existence of environment variables, and is not yet
// implemented in a truly re-windable cross-platform way, but it's a start, and it's better than asking someone over an email what failed
var commandLog = []string{}

func LogCommand(s string) {
	commandLog = append(commandLog, s)
}

func WriteCommandLogFile(error error) {
	//Write the entire command log to a file on the filesystem so that the user has the option of sending it to Gruntwork for debugging purposes
	// We currently write the file to ./gruntwork-examples-debug.log in the same directory as the executable was run in

	// Create the file
	f, err := os.Create("bocker-debug.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	// Write to the file, first writing the UTC timestamp as the first line, then looping through the command log to write each command on a new line
	f.WriteString("Ran at: " + time.Now().UTC().String() + "\n")
	f.WriteString("******************************************************************************\n")
	f.WriteString("Human legible log of steps taken and commands run up to the point of failure:\n")
	f.WriteString("******************************************************************************\n")
	for _, cmd := range commandLog {
		f.WriteString(cmd + "\n")
	}
	f.WriteString("^ The above command is likely the one that caused the error!\n")
	f.WriteString("\n\n")
	f.WriteString("******************************************************************************\n")
	f.WriteString("Complete log of the error that halted the deployment:\n")
	f.WriteString("******************************************************************************\n")
	f.WriteString("\n\n")
	f.WriteString(error.Error() + "\n")
}
