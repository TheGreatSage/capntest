module github.com/TheGreatSage/capntest

go 1.25.1

require (
	capnproto.org/go/capnp/v3 v3.1.0-alpha.1
	github.com/go-faker/faker/v4 v4.7.0
	google.golang.org/protobuf v1.36.10
	wellquite.org/bebop v0.0.0-20250203143624-8fd9a90b00fa
)

require (
	github.com/colega/zeropool v0.0.0-20230505084239-6fb4a4f75381 // indirect
	github.com/google/uuid v1.6.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/text v0.30.0 // indirect
)

//replace capnproto.org/go/capnp/v3 => ../../github/go-capnp

replace capnproto.org/go/capnp/v3 => github.com/TheGreatSage/go-capnp/v3 v3.1.2-sage
