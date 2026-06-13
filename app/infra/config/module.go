package config

import "go.uber.org/fx"

// Module exposes the infrastructure constructors to the Fx application.
var Module = fx.Module(
	"infrastructure",
	fx.Provide(
		NewViper,
		NewAppConfig,
		NewLogger,
		NewDB,
	),
)
