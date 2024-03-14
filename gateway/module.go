package gateway

import "go.uber.org/fx"

var HTTPServerModule = fx.Module("http_server",
	fx.Provide(
		New,
	),
)
