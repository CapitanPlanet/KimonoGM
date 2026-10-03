package main

import (
    "context"
    "embed"
)

//go:embed frontend/src/assets
//go:embed frontend/templates
var templatesFS embed.FS

type App struct {
    ctx         context.Context
    projectPath string
}

func NewApp() *App { return &App{} }
func (a *App) startup(ctx context.Context) { a.ctx = ctx }
