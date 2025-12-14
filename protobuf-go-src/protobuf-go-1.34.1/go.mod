module google.golang.org/protobuf

go 1.17

require (
	github.com/golang/protobuf v1.5.0
	github.com/google/go-cmp v0.5.5
)

require golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543 // indirect

replace github.com/golang/protobuf => ../deps/protobuf-1.5.0

replace github.com/google/go-cmp => ../deps/go-cmp-0.5.5

replace golang.org/x/xerrors => ../deps/xerrors-master
