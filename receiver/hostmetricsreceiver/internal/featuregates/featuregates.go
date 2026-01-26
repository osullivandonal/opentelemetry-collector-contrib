package featuregates

import "go.opentelemetry.io/collector/featuregate"

var EmitOldMetrics = featuregate.GlobalRegistry().MustRegister(
	"receiver.hostmetricsreceiver.emitOldMetrics",
	featuregate.StageBeta,
	featuregate.WithRegisterFromVersion("v1.9.0"),
	featuregate.WithRegisterDescription("Emit metrics using legacy schema"),
	featuregate.WithRegisterReferenceURL("https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/45592"),
)

var EmitSemconvMetrics = featuregate.GlobalRegistry().MustRegister(
	"receiver.hostmetricsreceiver.emitSemconvMetrics",
	featuregate.StageAlpha,
	featuregate.WithRegisterFromVersion("v1.39.0"),
	featuregate.WithRegisterDescription("Emit metrics using the semconv schema"),
	featuregate.WithRegisterReferenceURL("https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/45592"),
)
