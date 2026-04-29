//go:build windows

package psadt

import (
	"fmt"

	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/internal/parser"
	"github.com/pedrostefanogv/go-psadt/types"
)

// StartProcess starts an executable process.
func (s *Session) StartProcess(opts types.StartProcessOptions) (*types.ProcessResult, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("Start-ADTProcess", opts)
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	if !opts.PassThru {
		if err := parser.CheckSuccess(data); err != nil {
			return nil, err
		}
		return nil, nil
	}
	var result types.ProcessResult
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// StartProcessAsUser starts a process in the user's session.
func (s *Session) StartProcessAsUser(opts types.StartProcessAsUserOptions) (*types.ProcessResult, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("Start-ADTProcessAsUser", opts)
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	if !opts.PassThru {
		if err := parser.CheckSuccess(data); err != nil {
			return nil, err
		}
		return nil, nil
	}
	var result types.ProcessResult
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// StartMsiProcess starts an MSI installation/uninstallation process.
func (s *Session) StartMsiProcess(opts types.MsiProcessOptions) (*types.ProcessResult, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("Start-ADTMsiProcess", opts)
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	if !opts.PassThru {
		if err := parser.CheckSuccess(data); err != nil {
			return nil, err
		}
		return nil, nil
	}
	var result types.ProcessResult
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// StartMsiProcessAsUser starts an MSI process in the user's session.
func (s *Session) StartMsiProcessAsUser(opts types.MsiProcessAsUserOptions) (*types.ProcessResult, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("Start-ADTMsiProcessAsUser", opts)
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	if !opts.PassThru {
		if err := parser.CheckSuccess(data); err != nil {
			return nil, err
		}
		return nil, nil
	}
	var result types.ProcessResult
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// StartMspProcess starts an MSP (patch) process.
func (s *Session) StartMspProcess(opts types.MspProcessOptions) (*types.ProcessResult, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("Start-ADTMspProcess", opts)
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	if !opts.PassThru {
		if err := parser.CheckSuccess(data); err != nil {
			return nil, err
		}
		return nil, nil
	}
	var result types.ProcessResult
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// StartMspProcessAsUser starts an MSP (patch) process in the user's session.
func (s *Session) StartMspProcessAsUser(opts types.MspProcessAsUserOptions) (*types.ProcessResult, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("Start-ADTMspProcessAsUser", opts)
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	if !opts.PassThru {
		if err := parser.CheckSuccess(data); err != nil {
			return nil, err
		}
		return nil, nil
	}
	var result types.ProcessResult
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BlockAppExecution blocks execution of specified applications.
func (s *Session) BlockAppExecution(processes []types.ProcessDefinition, position ...types.DialogPosition) error {
	ctx, cancel := s.getContext()
	defer cancel()

	type blockOpts struct {
		ProcessName []types.ProcessDefinition `ps:"ProcessName"`
		Position    types.DialogPosition      `ps:"Position"`
	}
	opts := blockOpts{ProcessName: processes}
	if len(position) > 0 {
		opts.Position = position[0]
	}

	cmd := cmdbuilder.Build("Block-ADTAppExecution", opts)
	return s.executeVoid(ctx, cmd)
}

// UnblockAppExecution unblocks application execution.
func (s *Session) UnblockAppExecution() error {
	ctx, cancel := s.getContext()
	defer cancel()
	return s.executeVoid(ctx, "Unblock-ADTAppExecution")
}

// GetRunningProcesses checks for running processes from a list of names.
func (s *Session) GetRunningProcesses(processNames []string) ([]types.RunningProcess, error) {
	ctx, cancel := s.getContext()
	defer cancel()

	cmd := fmt.Sprintf("Get-ADTRunningProcesses -ProcessName %s", cmdbuilder.EscapeArray(processNames))
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var result []types.RunningProcess
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}
