package models

// Result holds the output of a single command executed on a device.

type Result struct {
	Cmd    string
	Output string
}