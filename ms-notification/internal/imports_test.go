package internal_test

import (
	"testing"

	"github.com/matthewmcnew/archtest"
)

func Test_Service_ShouldNotDependOn_Consumer(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "ms-notification/internal/service").ShouldNotDependOn(
		"ms-notification/internal/consumer",
	)
}

func Test_Domain_ShouldNotDependOn_Consumer(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "ms-notification/internal/domain").ShouldNotDependOn(
		"ms-notification/internal/consumer",
	)
}

func Test_Domain_ShouldNotDependOn_Service(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "ms-notification/internal/domain").ShouldNotDependOn(
		"ms-notification/internal/service",
	)
}

func Test_Pkg_ShouldNotDependOn_Consumer(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "ms-notification/pkg/...").ShouldNotDependOn(
		"ms-notification/internal/consumer",
	)
}

func Test_Pkg_ShouldNotDependOn_Service(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "ms-notification/pkg/...").ShouldNotDependOn(
		"ms-notification/internal/service",
	)
}
