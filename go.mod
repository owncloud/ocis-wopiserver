module github.com/owncloud/ocis-wopiserver

go 1.16

require (
	contrib.go.opencensus.io/exporter/jaeger v0.2.1
	contrib.go.opencensus.io/exporter/ocagent v0.7.0
	contrib.go.opencensus.io/exporter/zipkin v0.1.2
	github.com/asim/go-micro/v3 v3.6.0
	github.com/cs3org/go-cs3apis v0.0.0-20210812121411-f18cf19614e8
	github.com/cs3org/reva v1.12.0
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-chi/chi/v5 v5.0.4
	github.com/micro/cli/v2 v2.1.2
	github.com/oklog/run v1.1.0
	github.com/openzipkin/zipkin-go v0.2.5
	github.com/owncloud/ocis v1.11.0
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/viper v1.8.1
	go.opencensus.io v0.23.0
	golang.org/x/net v0.0.0-20210825183410-e898025ed96a
	google.golang.org/grpc v1.40.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
)

replace (
	github.com/crewjam/saml => github.com/crewjam/saml v0.4.5
	go.etcd.io/etcd/api/v3 => go.etcd.io/etcd/api/v3 v3.0.0-20210204162551-dae29bb719dd
	go.etcd.io/etcd/pkg/v3 => go.etcd.io/etcd/pkg/v3 v3.0.0-20210204162551-dae29bb719dd
)
