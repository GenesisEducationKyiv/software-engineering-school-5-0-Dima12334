package internal_test

import (
	"testing"

	"github.com/matthewmcnew/archtest"
)

func Test_Service_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "ms-weather-subscription/weather_forecast_sub/internal/service").ShouldNotDependOn(
		"ms-weather-subscription/weather_forecast_sub/internal/handlers",
	)
}

func Test_Repository_ShouldNotDependOn_Service(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "ms-weather-subscription/weather_forecast_sub/internal/repository").ShouldNotDependOn(
		"ms-weather-subscription/weather_forecast_sub/internal/service",
	)
}

func Test_Repository_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "ms-weather-subscription/weather_forecast_sub/internal/repository").ShouldNotDependOn(
		"ms-weather-subscription/weather_forecast_sub/internal/handlers",
	)
}

func Test_Domain_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "ms-weather-subscription/weather_forecast_sub/internal/domain").ShouldNotDependOn(
		"ms-weather-subscription/weather_forecast_sub/internal/handlers",
	)
}

func Test_Domain_ShouldNotDependOn_Service(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "ms-weather-subscription/weather_forecast_sub/internal/domain").ShouldNotDependOn(
		"ms-weather-subscription/weather_forecast_sub/internal/service",
	)
}

func Test_Domain_ShouldNotDependOn_Repository(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "ms-weather-subscription/weather_forecast_sub/internal/domain").ShouldNotDependOn(
		"ms-weather-subscription/weather_forecast_sub/internal/repository",
	)
}

func Test_Pkg_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "ms-weather-subscription/weather_forecast_sub/pkg/...").ShouldNotDependOn(
		"ms-weather-subscription/weather_forecast_sub/internal/handlers",
	)
}

func Test_Pkg_ShouldNotDependOn_Service(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "ms-weather-subscription/weather_forecast_sub/pkg/...").ShouldNotDependOn(
		"ms-weather-subscription/weather_forecast_sub/internal/service",
	)
}

func Test_Pkg_ShouldNotDependOn_Repository(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "ms-weather-subscription/weather_forecast_sub/pkg/...").ShouldNotDependOn(
		"ms-weather-subscription/weather_forecast_sub/internal/repository",
	)
}
