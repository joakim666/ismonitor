package main

import (
	"fmt"
	"strconv"
	"strings"
)

func verifyFreeSpace(output string, diskUsagePercentWarning int) []string {
	var errors []string

	lines := strings.Split(output, "\n")

	for _, line := range lines[1:] { // skip header row
		columns := strings.Fields(line)

		if len(columns) > 1 {
			percentValue, err := strconv.Atoi(strings.Trim(columns[1], "%"))
			if err != nil {
				errors = append(errors, fmt.Sprint(err))
			} else {
				if percentValue >= diskUsagePercentWarning {
					errors = append(errors, fmt.Sprintf("Disk usage at %d percent\n", percentValue))
				}
			}
		}
	}

	return errors
}

func verifyLoadAvg(output string, uptimeLoad5MinutesWarning float64) []string {
	var errors []string

	columns := strings.Fields(output)

	if len(columns) >= 3 {
		floatValue, err := strconv.ParseFloat(columns[1], 64)
		if err != nil {
			errors = append(errors, fmt.Sprint(err))
		} else {
			if floatValue >= uptimeLoad5MinutesWarning {
				errors = append(errors, fmt.Sprintf("High load warning: %s\n", output))
			}
		}
	}

	return errors
}
