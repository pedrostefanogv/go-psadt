//go:build windows

package types

import (
	"testing"
)

// TestEnumConstantsUniqueAndNonEmpty guards against typos in enum values:
// every constant in a set must be non-empty and unique within the set.
func TestEnumConstantsUniqueAndNonEmpty(t *testing.T) {
	checks := map[string][]string{
		"DeploymentType":        {string(DeployInstall), string(DeployUninstall), string(DeployRepair)},
		"DeployMode":            {string(DeployModeAuto), string(DeployModeInteractive), string(DeployModeNonInteractive), string(DeployModeSilent)},
		"DialogStyle":           {string(DialogStyleFluent), string(DialogStyleClassic)},
		"DialogPosition":        {string(DialogPositionDefault), string(DialogPositionTopLeft), string(DialogPositionTop), string(DialogPositionTopRight), string(DialogPositionTopCenter), string(DialogPositionCenter), string(DialogPositionBottomLeft), string(DialogPositionBottom), string(DialogPositionBottomRight), string(DialogPositionBottomCenter), string(DialogPositionOobe)},
		"DialogSystemIcon":      {string(IconApplication), string(IconAsterisk), string(IconError), string(IconExclamation), string(IconHand), string(IconInformation), string(IconQuestion), string(IconShield), string(IconWarning), string(IconWinLogo)},
		"DialogBoxButtons":      {string(ButtonsOk), string(ButtonsOkCancel), string(ButtonsAbortRetryIgnore), string(ButtonsYesNoCancel), string(ButtonsYesNo), string(ButtonsRetryCancel), string(ButtonsCancelTryContinue)},
		"RegistryValueKind":     {string(RegString), string(RegExpandString), string(RegBinary), string(RegDWord), string(RegMultiString), string(RegQWord)},
		"ProcessWindowStyle":    {string(WindowNormal), string(WindowHidden), string(WindowMaximized), string(WindowMinimized)},
		"EnvironmentVarTarget":  {string(EnvTargetProcess), string(EnvTargetUser), string(EnvTargetMachine)},
		"ServiceStartMode":      {string(ServiceAutomatic), string(ServiceManual), string(ServiceDisabled), string(ServiceAutomaticDelayedStart)},
		"MsiAction":             {string(MsiInstall), string(MsiUninstall), string(MsiPatch), string(MsiRepair), string(MsiActiveSetup)},
		"NameMatch":             {string(MatchContains), string(MatchExact), string(MatchWildcard), string(MatchRegex)},
		"ApplicationType":       {string(AppTypeAll), string(AppTypeMSI), string(AppTypeEXE)},
		"BalloonTipIcon":        {string(BalloonNone), string(BalloonInfo), string(BalloonWarning), string(BalloonError)},
		"DialogBoxDefaultBtn":   {string(DialogDefaultFirst), string(DialogDefaultSecond), string(DialogDefaultThird)},
		"MessageAlignment":      {string(AlignLeft), string(AlignCenter), string(AlignRight)},
	}

	for set, values := range checks {
		seen := map[string]bool{}
		for _, v := range values {
			if v == "" {
				t.Errorf("%s: empty enum constant", set)
			}
			if seen[v] {
				t.Errorf("%s: duplicated enum value %q", set, v)
			}
			seen[v] = true
		}
	}
}

func TestEnumValidMethods(t *testing.T) {
	if !(DeployInstall).Valid() || !(DeployUninstall).Valid() || !(DeployRepair).Valid() {
		t.Error("all DeploymentType constants must be valid")
	}
	if (DeploymentType("Bogus")).Valid() {
		t.Error("bogus DeploymentType must be invalid")
	}
	if !(DeployMode("")).Valid() {
		t.Error("empty DeployMode must be treated as valid (Auto)")
	}
	if !(MsiPatch).Valid() || (MsiAction("X")).Valid() {
		t.Error("MsiAction.Valid is wrong")
	}
	if !(WindowHidden).Valid() || !(ProcessWindowStyle("")).Valid() || (ProcessWindowStyle("Big")).Valid() {
		t.Error("ProcessWindowStyle.Valid is wrong")
	}
}
