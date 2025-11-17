module example

go 1.23.9

replace github.com/ruko1202/xlog => ./../

require (
	github.com/google/uuid v1.6.0
	github.com/ruko1202/xlog v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.27.0
)

require go.uber.org/multierr v1.11.0 // indirect
