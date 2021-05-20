module github.com/owncloud/ocis-wopiserver

go 1.16

require (
	contrib.go.opencensus.io/exporter/jaeger v0.2.1
	contrib.go.opencensus.io/exporter/ocagent v0.7.0
	contrib.go.opencensus.io/exporter/zipkin v0.1.2
	github.com/asim/go-micro/v3 v3.5.1-0.20210217182006-0f0ace1a44a9
	github.com/cs3org/go-cs3apis v0.0.0-20210507060801-f176760d55f4
	github.com/cs3org/reva v1.7.1-0.20210507160327-e2c3841d0dbc
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/micro/cli/v2 v2.1.2
	github.com/oklog/run v1.1.0
	github.com/openzipkin/zipkin-go v0.2.5
	github.com/owncloud/ocis/ocis-pkg v0.0.0-20210519113029-34a8ed381620
	github.com/prometheus/client_golang v1.10.0
	github.com/spf13/viper v1.7.1
	go.opencensus.io v0.23.0
	golang.org/x/net v0.0.0-20210428140749-89ef3d95e781
	google.golang.org/genproto v0.0.0-20210413151531-c14fb6ef47c3 // indirect
	google.golang.org/grpc v1.37.0
)

replace (
	github.com/gomodule/redigo => github.com/gomodule/redigo v1.8.2
	github.com/oleiade/reflections => github.com/oleiade/reflections v1.0.1
	// taken from https://github.com/asim/go-micro/blob/master/plugins/registry/etcd/go.mod#L14-L16
	go.etcd.io/etcd/api/v3 => go.etcd.io/etcd/api/v3 v3.0.0-20210204162551-dae29bb719dd
	go.etcd.io/etcd/pkg/v3 => go.etcd.io/etcd/pkg/v3 v3.0.0-20210204162551-dae29bb719dd
	// latest version compatible with etcd
	google.golang.org/grpc => google.golang.org/grpc v1.29.1
)
