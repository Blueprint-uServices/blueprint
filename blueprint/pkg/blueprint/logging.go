package blueprint

import (
	"golang.org/x/exp/slog"
)

func InitBlueprintCompilerLogging() {
	// // For now don't bother configuring logger.  Maybe in future

	// var programLevel = new(slog.LevelVar)
	// programLevel.Set(slog.LevelError)

	// opts := slog.HandlerOptions{
	// 	Level:     programLevel,
	// 	AddSource: true,
	// }

	// logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	// slog.SetDefault(logger)
	slog.Info("Hello world")

}
