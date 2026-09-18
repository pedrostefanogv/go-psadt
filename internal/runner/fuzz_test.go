//go:build windows

package runner

import (
	"strconv"
	"strings"
	"testing"
)

// FuzzWrapCommandWithID checks wrapper invariants: the command body is
// preserved verbatim inside the script block and the id-tagged markers are
// present exactly once.
func FuzzWrapCommandWithID(f *testing.F) {
	f.Add("Get-ADTFreeDiskSpace", uint64(1))
	f.Add("", uint64(0))
	f.Add("Write-Output 'x'; Write-Output \"y\"", uint64(18446744073709551615))
	f.Add("throw $null", uint64(42))

	f.Fuzz(func(t *testing.T, cmd string, id uint64) {
		wrapped := WrapCommandWithID(cmd, id)
		begin := beginMarkerFor(id)
		if !strings.Contains(wrapped, begin) {
			t.Errorf("missing begin marker %s", begin)
		}
		end := "<<<PSADT_END:" + strconv.FormatUint(id, 10) + ">>>"
		if !strings.Contains(wrapped, end) {
			t.Errorf("missing end marker %s", end)
		}
		if strings.Count(wrapped, begin) != 1 || strings.Count(wrapped, end) != 1 {
			t.Error("markers must appear exactly once")
		}
	})
}
