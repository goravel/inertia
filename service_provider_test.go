package inertia

import (
	"io"
	"testing"

	"github.com/goravel/framework/contracts/foundation"
	contractsvite "github.com/goravel/framework/contracts/vite"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	"github.com/goravel/framework/support/color"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mocksinertia "github.com/goravel/inertia/mocks"
)

func TestServiceProviderRegistersViteFuncs(t *testing.T) {
	app := mocksfoundation.NewApplication(t)
	config := mocksconfig.NewConfig(t)
	config.EXPECT().GetString("inertia.root_view", "app").Return("app").Once()
	config.EXPECT().GetString("inertia.version", "").Return("").Once()
	config.EXPECT().GetBool("inertia.ssr", false).Return(false).Once()
	config.EXPECT().GetString("inertia.ssr_url", "http://127.0.0.1:13714/render").Return("http://127.0.0.1:13714/render").Once()
	config.EXPECT().GetInt("inertia.ssr_timeout", 5).Return(5).Once()
	config.EXPECT().GetString("app.url", "").Return("").Once()
	config.EXPECT().Get("inertia.flash_keys").Return(nil).Once()

	app.EXPECT().MakeConfig().Return(config).Once()
	app.EXPECT().MakeVite().Return(newTestService(t, `{"resources/js/app.ts":{"file":"assets/app-abc.js"}}`)).Once()
	app.EXPECT().Singleton("goravel.inertia", mock.Anything).Run(func(_ any, callback func(app foundation.Application) (any, error)) {
		instance, err := callback(app)
		require.NoError(t, err)
		require.IsType(t, &InertiaManager{}, instance)
		// The manifest hash is the version the client compares on every visit.
		require.NotEmpty(t, instance.(*InertiaManager).Version())
	}).Once()

	(&ServiceProvider{}).Register(app)
}

// expectBoot sets the expectations Boot needs besides MakeVite, so a test can focus
// on the warning it emits.
func expectBoot(t *testing.T, app *mocksfoundation.Application, vite contractsvite.Vite) {
	t.Helper()

	config := mocksconfig.NewConfig(t)
	artisan := mocksconsole.NewArtisan(t)
	manager := mocksinertia.NewInertia(t)

	app.EXPECT().MakeVite().Return(vite).Once()
	app.EXPECT().Make("goravel.inertia").Return(manager, nil).Once()
	app.EXPECT().MakeConfig().Return(config).Once()
	app.EXPECT().MakeArtisan().Return(artisan).Once()
	config.EXPECT().GetString("app.name").Return("Goravel").Once()
	manager.EXPECT().Share("appName", "Goravel").Once()
	manager.EXPECT().ShareFunc("timestamp", mock.Anything).Once()
	artisan.EXPECT().Register(mock.Anything).Once()
}

func TestBootWarnsWhenViteProviderMissing(t *testing.T) {
	missing := mocksfoundation.NewApplication(t)
	expectBoot(t, missing, nil)

	output := color.CaptureOutput(func(io.Writer) { (&ServiceProvider{}).Boot(missing) })
	require.Contains(t, output, "&vite.ServiceProvider{}")

	bound := mocksfoundation.NewApplication(t)
	expectBoot(t, bound, newTestService(t, ""))

	require.Empty(t, color.CaptureOutput(func(io.Writer) { (&ServiceProvider{}).Boot(bound) }))
}
