package inertia

import (
	"io"
	"strings"
	"testing"

	"github.com/goravel/framework/contracts/foundation"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	"github.com/goravel/framework/support/color"
	frameworkvite "github.com/goravel/framework/vite"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func TestWarnIfViteMissing(t *testing.T) {
	app := mocksfoundation.NewApplication(t)
	app.EXPECT().MakeVite().Return(nil).Once()
	output := color.CaptureOutput(func(io.Writer) { warnIfViteMissing(app) })
	if !strings.Contains(output, "&vite.ServiceProvider{}") {
		t.Errorf("warning should name the provider to register, got: %q", output)
	}

	bound := mocksfoundation.NewApplication(t)
	bound.EXPECT().MakeVite().Return(frameworkvite.NewVite(nil)).Once()
	if output := color.CaptureOutput(func(io.Writer) { warnIfViteMissing(bound) }); output != "" {
		t.Errorf("no warning expected when vite is bound, got: %q", output)
	}
}
