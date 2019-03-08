package ui

// PrintProfile specifies the printing mode
type PrintProfile int

// enumeration of print profile types
const (
	// the following profiles are generated as a result of the emulation
	CPUStep PrintProfile = iota
	VideoStep
	MachineInfo
	MachineInfoInternal

	// the following profiles are generated by the debugger in response to user
	// input
	Feedback
	Prompt
	Script
	Help

	// user input (not used by all user interface types [eg. echoing terminals])
	Input

	// errors can be generated by the emulation or the debugger
	Error
)
