package internal_test

import (
	"testing"

	"github.com/matthewmcnew/archtest"
)

func Test_Service_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/internal/service").ShouldNotDependOn(
		"weather_forecast_sub/internal/handlers",
	)
}

func Test_Repository_ShouldNotDependOn_Service(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/internal/repository").ShouldNotDependOn(
		"weather_forecast_sub/internal/service",
	)
}

func Test_Repository_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/internal/repository").ShouldNotDependOn(
		"weather_forecast_sub/internal/handlers",
	)
}

func Test_Domain_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/internal/domain").ShouldNotDependOn(
		"weather_forecast_sub/internal/handlers",
	)
}

func Test_Domain_ShouldNotDependOn_Service(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/internal/domain").ShouldNotDependOn(
		"weather_forecast_sub/internal/service",
	)
}

func Test_Domain_ShouldNotDependOn_Repository(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/internal/domain").ShouldNotDependOn(
		"weather_forecast_sub/internal/repository",
	)
}

func Test_PkgCache_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/cache").ShouldNotDependOn(
		"weather_forecast_sub/internal/handlers",
	)
}

func Test_PkgCache_ShouldNotDependOn_Service(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/cache").ShouldNotDependOn(
		"weather_forecast_sub/internal/service",
	)
}

func Test_PkgCache_ShouldNotDependOn_Repository(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/cache").ShouldNotDependOn(
		"weather_forecast_sub/internal/repository",
	)
}

func Test_PkgClients_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/clients").ShouldNotDependOn(
		"weather_forecast_sub/internal/handlers",
	)
}

func Test_PkgClients_ShouldNotDependOn_Service(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/clients").ShouldNotDependOn(
		"weather_forecast_sub/internal/service",
	)
}

func Test_PkgClients_ShouldNotDependOn_Repository(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/clients").ShouldNotDependOn(
		"weather_forecast_sub/internal/repository",
	)
}

func Test_PkgEmail_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/email").ShouldNotDependOn(
		"weather_forecast_sub/internal/handlers",
	)
}

func Test_PkgEmail_ShouldNotDependOn_Service(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/email").ShouldNotDependOn(
		"weather_forecast_sub/internal/service",
	)
}

func Test_PkgEmail_ShouldNotDependOn_Repository(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/email").ShouldNotDependOn(
		"weather_forecast_sub/internal/repository",
	)
}

func Test_PkgSMPT_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/email/smtp").ShouldNotDependOn(
		"weather_forecast_sub/internal/handlers",
	)
}

func Test_PkgSMPT_ShouldNotDependOn_Service(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/email/smtp").ShouldNotDependOn(
		"weather_forecast_sub/internal/service",
	)
}

func Test_PkgSMPT_ShouldNotDependOn_Repository(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/email/smtp").ShouldNotDependOn(
		"weather_forecast_sub/internal/repository",
	)
}

func Test_PkgErrors_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/errors").ShouldNotDependOn(
		"weather_forecast_sub/internal/handlers",
	)
}

func Test_PkgErrors_ShouldNotDependOn_Service(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/errors").ShouldNotDependOn(
		"weather_forecast_sub/internal/service",
	)
}

func Test_PkgErrors_ShouldNotDependOn_Repository(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/errors").ShouldNotDependOn(
		"weather_forecast_sub/internal/repository",
	)
}

func Test_PkgHash_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/hash").ShouldNotDependOn(
		"weather_forecast_sub/internal/handlers",
	)
}

func Test_PkgHash_ShouldNotDependOn_Service(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/hash").ShouldNotDependOn(
		"weather_forecast_sub/internal/service",
	)
}

func Test_PkgHash_ShouldNotDependOn_Repository(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/hash").ShouldNotDependOn(
		"weather_forecast_sub/internal/repository",
	)
}

func Test_PkgLogger_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/logger").ShouldNotDependOn(
		"weather_forecast_sub/internal/handlers",
	)
}

func Test_PkgLogger_ShouldNotDependOn_Service(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/logger").ShouldNotDependOn(
		"weather_forecast_sub/internal/service",
	)
}

func Test_PkgLogger_ShouldNotDependOn_Repository(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/logger").ShouldNotDependOn(
		"weather_forecast_sub/internal/repository",
	)
}

func Test_PkgMigrations_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/migrations").ShouldNotDependOn(
		"weather_forecast_sub/internal/handlers",
	)
}

func Test_PkgMigrations_ShouldNotDependOn_Service(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/migrations").ShouldNotDependOn(
		"weather_forecast_sub/internal/service",
	)
}

func Test_PkgMigrations_ShouldNotDependOn_Repository(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather_forecast_sub/pkg/migrations").ShouldNotDependOn(
		"weather_forecast_sub/internal/repository",
	)
}
