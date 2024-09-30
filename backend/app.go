package backend

import (
	"context"
	"ftb-debug-ui/backend/dbg"
	"ftb-debug-ui/backend/fixes"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) RunDebug() (string, error) {
	return dbg.RunDebug()
}

func (a *App) RunCommonFixes() error {
	return fixes.FixCommonIssues()
}
