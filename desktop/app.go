package main

import (
	"context"

	"github.com/ankit-lilly/nqcli/internal/core"
	"github.com/ankit-lilly/nqcli/internal/desktop"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func runApp(svc core.QueryService, profile string, factory desktop.ServiceFactory) error {
	desktopSvc := desktop.NewDesktopService(svc, profile, factory)

	app := application.New(application.Options{
		Name: "nq",
		Services: []application.Service{
			application.NewService(&desktopLifecycle{service: desktopSvc}),
			application.NewService(desktopSvc),
		},
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:          "dGrapher",
		Width:          1440,
		Height:         900,
		URL:            "/",
		Frameless:      true,
		BackgroundType: application.BackgroundTypeTranslucent,
		Mac: application.MacWindow{
			Backdrop: application.MacBackdropTranslucent,
			TitleBar: application.MacTitleBar{
				AppearsTransparent: true,
			},
			InvisibleTitleBarHeight: 40,
		},
	})

	return app.Run()
}

// Keep platform dependencies in the desktop entry point so the service remains
// testable without WebKit/GTK and usable by non-GUI integration tests.
type desktopLifecycle struct{ service *desktop.DesktopService }

var _ application.ServiceStartup = (*desktopLifecycle)(nil)
var _ application.ServiceShutdown = (*desktopLifecycle)(nil)

func (d *desktopLifecycle) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	return d.service.Startup(ctx)
}
func (d *desktopLifecycle) ServiceShutdown() error { return d.service.Shutdown() }
