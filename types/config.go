//go:build windows

package types

// ADTConfig represents the PSADT configuration returned by Get-ADTConfig.
type ADTConfig struct {
	Toolkit map[string]interface{} `json:"Toolkit"`
	UI      map[string]interface{} `json:"UI"`
	MSI     map[string]interface{} `json:"MSI"`
}

// ADTStringTable represents the PSADT string table returned by Get-ADTStringTable.
type ADTStringTable struct {
	Messages map[string]string `json:"Messages"`
}

// DeferHistory represents defer history.
type DeferHistory struct {
	DeferTimesRemaining int    `json:"DeferTimesRemaining"`
	DeferDeadline       string `json:"DeferDeadline"`
}

// SendKeysOptions options for Send-ADTKeys.
type SendKeysOptions struct {
	WindowTitle string `ps:"WindowTitle"`
	Keys        string `ps:"Keys"`
	WaitSeconds int    `ps:"WaitSeconds"`
}

// RetryOptions options for Invoke-ADTCommandWithRetries.
type RetryOptions struct {
	Command      string `ps:"Command"`
	MaxRetries   int    `ps:"MaxRetries"`
	SleepSeconds int    `ps:"SleepSeconds"`
}

// SCCMTaskOptions options for Invoke-ADTSCCMTask.
type SCCMTaskOptions struct {
	ScheduleID string `ps:"ScheduleID"`
}
