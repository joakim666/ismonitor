package main

import (
	"fmt"
	"strconv"
	"strings"
)

func verifyFreeSpace(output string, diskUsagePercentWarning int) []verificationError {
	var errors []verificationError

	lines := strings.Split(output, "\n")

	for _, line := range lines[1:] { // skip header row
		columns := strings.Fields(line)

		if len(columns) > 1 {
			percentValue, err := strconv.Atoi(strings.Trim(columns[1], "%"))
			if err != nil {
				e := verificationError{title: "Disk usage verification error", message: fmt.Sprint(err)}
				errors = append(errors, e)
			} else {
				if percentValue >= diskUsagePercentWarning {
					e := verificationError{title: "Disk usage verification error", message: fmt.Sprintf("Disk usage at %d percent\n", percentValue)}
					errors = append(errors, e)
				}
			}
		}
	}

	return errors
}

func verifyLoadAvg(output string, uptimeLoad5MinutesWarning float64) []verificationError {
	var errors []verificationError

	columns := strings.Fields(output)

	if len(columns) >= 3 {
		floatValue, err := strconv.ParseFloat(columns[1], 64)
		if err != nil {
			e := verificationError{title: "Load average verification error", message: fmt.Sprint(err)}
			errors = append(errors, e)
		} else {
			if floatValue >= uptimeLoad5MinutesWarning {
				e := verificationError{title: "Load average verification error", message: fmt.Sprintf("High load warning: %s\n", output)}
				errors = append(errors, e)
			}
		}
	}

	return errors
}
