package main

import (
	"context"
	"sync"

	"github.com/ankit-lilly/nqcli/internal/core"
	"github.com/ankit-lilly/nqcli/internal/desktop"
	graphschema "github.com/ankit-lilly/nqcli/internal/schema"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const schemaEventName = "schema:update"

func init() {
	application.RegisterEvent[graphschema.Event](schemaEventName)
}

func runApp(svc core.QueryService, profile string, factory desktop.ServiceFactory) error {
	desktopSvc := desktop.NewDesktopService(svc, profile, factory)
	lifecycle := &desktopLifecycle{service: desktopSvc}

	app := application.New(application.Options{
		Name: "nq",
		Services: []application.Service{
			application.NewService(lifecycle),
			application.NewService(desktopSvc),
		},
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})
	lifecycle.emitSchemaEvent = func(event graphschema.Event) {
		app.Event.Emit(schemaEventName, event)
	}

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:           "dGrapher",
		Width:           1440,
		Height:          900,
		MinWidth:        900,
		MinHeight:       600,
		InitialPosition: application.WindowCentered,
		URL:             "/",
		Frameless:       true,
		BackgroundType:  application.BackgroundTypeTranslucent,
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
type desktopLifecycle struct {
	service         *desktop.DesktopService
	emitSchemaEvent func(graphschema.Event)
	unsubscribe     func()
	workers         sync.WaitGroup
}

var _ application.ServiceStartup = (*desktopLifecycle)(nil)
var _ application.ServiceShutdown = (*desktopLifecycle)(nil)

func (d *desktopLifecycle) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	if err := d.service.Startup(ctx); err != nil {
		return err
	}
	events, unsubscribe := d.service.SubscribeSchemaEvents()
	d.unsubscribe = unsubscribe
	d.workers.Go(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event, open := <-events:
				if !open {
					return
				}
				if d.emitSchemaEvent != nil {
					d.emitSchemaEvent(event)
				}
			}
		}
	})
	return nil
}
func (d *desktopLifecycle) ServiceShutdown() error {
	if d.unsubscribe != nil {
		d.unsubscribe()
	}
	err := d.service.Shutdown()
	d.workers.Wait()
	return err
}
