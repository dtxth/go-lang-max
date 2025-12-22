module gateway-service

go 1.24.0

require (
	auth-service v0.0.0
	chat-service v0.0.0
	employee-service v0.0.0-00010101000000-000000000000
	github.com/leanovate/gopter v0.2.11
	github.com/stretchr/testify v1.11.1
	google.golang.org/grpc v1.77.0
	google.golang.org/protobuf v1.36.10
	structure-service v0.0.0-00010101000000-000000000000
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace auth-service => ../auth-service

replace chat-service => ../chat-service

replace employee-service => ../employee-service

replace structure-service => ../structure-service
