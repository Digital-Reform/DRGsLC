package droverride_test

import (
	"log/slog"
	"os"
	"testing"

	droverride "github.com/Digital-Reform/DRGsLC/DROverride"
)

func TestColoredText(t *testing.T) {
	slog.SetDefault(slog.New(droverride.NewDebugHandler(os.Stdout, &droverride.Options{Level: slog.LevelDebug})))

	slog.Debug("Debug Test")
	slog.Info("Info Test")
	slog.Warn("Warn Test")
	slog.Error("Error Test")
}

func TestAttribGroups(t *testing.T) {
	dslog := slog.New(droverride.NewDebugHandler(os.Stdout, &droverride.Options{Level: slog.LevelDebug}))

	dslog.Info("Info String", "Test Attribute", 24)
	dslog.WithGroup("Test Group").Info("Test group info log", "Attrib", 65)
}
