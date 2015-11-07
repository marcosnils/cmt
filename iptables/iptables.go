package iptables

import "strings"

// This code can be improved a lot!
// Please don't consider this production ready!
func Diff(a, b string) []string {
	alines := strings.Split(a, "\n")
	blines := strings.Split(b, "\n")

	var diff []string

	for _, aline := range alines {
		match := false
		for _, bline := range blines {
			if aline == bline {
				match = true
				break
			}
		}
		if !match {
			diff = append(diff, aline)
		}
	}

	return diff
}
