//go:build windows

package psadt

import (
	"fmt"

	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/internal/parser"
	"github.com/pedrostefanogv/go-psadt/types"
)

// TestBattery tests if the system is running on battery power.
func (s *Session) TestBattery() (*types.BatteryInfo, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	data, err := s.execute(ctx, "Test-ADTBattery")
	if err != nil {
		return nil, err
	}
	var result types.BatteryInfo
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TestCallerIsAdmin checks if the current user has admin privileges.
func (s *Session) TestCallerIsAdmin() (bool, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	data, err := s.execute(ctx, "Test-ADTCallerIsAdmin")
	if err != nil {
		return false, err
	}
	return parser.ParseBool(data)
}

// TestNetworkConnection checks if there is an active network connection.
func (s *Session) TestNetworkConnection() (bool, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	data, err := s.execute(ctx, "Test-ADTNetworkConnection")
	if err != nil {
		return false, err
	}
	return parser.ParseBool(data)
}

// TestMutexAvailability tests if a named mutex is available.
func (s *Session) TestMutexAvailability(mutexName string) (bool, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := fmt.Sprintf("Test-ADTMutexAvailability -MutexName %s", cmdbuilder.EscapeString(mutexName))
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return false, err
	}
	return parser.ParseBool(data)
}

// TestPowerPoint checks if PowerPoint is running in presentation mode.
func (s *Session) TestPowerPoint() (bool, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	data, err := s.execute(ctx, "Test-ADTPowerPoint")
	if err != nil {
		return false, err
	}
	return parser.ParseBool(data)
}

// TestMicrophoneInUse checks if a microphone is currently in use.
func (s *Session) TestMicrophoneInUse() (bool, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	data, err := s.execute(ctx, "Test-ADTMicrophoneInUse")
	if err != nil {
		return false, err
	}
	return parser.ParseBool(data)
}

// TestUserIsBusy checks if the user is busy (presentation mode, microphone in use, etc.).
func (s *Session) TestUserIsBusy() (bool, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	data, err := s.execute(ctx, "Test-ADTUserIsBusy")
	if err != nil {
		return false, err
	}
	return parser.ParseBool(data)
}

// TestEspActive checks if the Enrollment Status Page is active.
func (s *Session) TestEspActive() (bool, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	data, err := s.execute(ctx, "Test-ADTEspActive")
	if err != nil {
		return false, err
	}
	return parser.ParseBool(data)
}

// TestOobeCompleted checks if the Out-Of-Box Experience has been completed.
func (s *Session) TestOobeCompleted() (bool, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	data, err := s.execute(ctx, "Test-ADTOobeCompleted")
	if err != nil {
		return false, err
	}
	return parser.ParseBool(data)
}

// TestMSUpdates checks if Microsoft Updates are available.
func (s *Session) TestMSUpdates() (bool, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	data, err := s.execute(ctx, "Test-ADTMSUpdates")
	if err != nil {
		return false, err
	}
	return parser.ParseBool(data)
}

// Ensure unused imports don't cause errors
var _ = cmdbuilder.EscapeString
