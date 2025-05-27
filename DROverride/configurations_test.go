package droverride_test

import (
	"log/slog"
	"os"
	"testing"

	droverride "github.com/Digital-Reform/DRGsLC/DROverride"
)

func TestBattery(t *testing.T) {
	slog.SetDefault(slog.New(droverride.NewDebugHandler(os.Stdout, &droverride.Options{Level: slog.LevelDebug})))

	slog.Debug("Debug Test")
	slog.Info("Info Test")
	slog.Warn("Warn Test")
	slog.Error("Error Test")
}
