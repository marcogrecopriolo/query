module github.com/couchbase/query

go 1.18

replace golang.org/x/text => golang.org/x/text v0.4.0

replace github.com/couchbase/cbauth => ../cbauth

replace github.com/couchbase/cbft => ../../../../../cbft

replace github.com/couchbase/cbftx => ../../../../../cbftx

replace github.com/couchbase/hebrew => ../../../../../hebrew

replace github.com/couchbase/cbgt => ../../../../../cbgt

replace github.com/couchbase/eventing-ee => ../eventing-ee

replace github.com/couchbase/go-couchbase => ../go-couchbase

replace github.com/couchbase/go_json => ../go_json

replace github.com/couchbase/gomemcached => ../gomemcached

replace github.com/couchbase/goutils => ../goutils

replace github.com/couchbase/godbc => ../godbc

replace github.com/couchbase/indexing => ../indexing

replace github.com/couchbase/gometa => ../gometa

replace github.com/couchbase/n1fty => ../n1fty

replace github.com/couchbase/plasma => ../plasma

replace github.com/couchbase/query => ./empty

replace github.com/couchbase/query-ee => ../query-ee

replace github.com/couchbase/regulator => ../regulator

require (
	github.com/couchbase/cbauth v0.1.10
	github.com/couchbase/clog v0.1.0
	github.com/couchbase/eventing-ee v0.0.0-00010101000000-000000000000
	github.com/couchbase/go-couchbase v0.1.1
	github.com/couchbase/go_json v0.0.0-20220330123059-4473a21887c8
	github.com/couchbase/gocbcore/v10 v10.2.6-0.20230628164442-47b8f45095ec
	github.com/couchbase/godbc v0.0.0-20210615212222-79da1b49cb4d
	github.com/couchbase/gomemcached v0.2.2-0.20230407174933-7d7ce13da8cc
	github.com/couchbase/goutils v0.1.2
	github.com/couchbase/indexing v0.0.0-20220923223016-071e9308c875
	github.com/couchbase/n1fty v0.0.0-00010101000000-000000000000
	github.com/couchbase/query-ee v0.0.0-00010101000000-000000000000
	github.com/couchbase/regulator v0.0.0-00010101000000-000000000000
	github.com/couchbasedeps/go-curl v0.0.0-20190830233031-f0b2afc926ec
	github.com/golang/snappy v0.0.4
	github.com/kylelemons/godebug v1.1.0
	github.com/mattn/go-runewidth v0.0.3
	github.com/peterh/liner v1.2.0
	github.com/russross/blackfriday v1.5.2
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414
	golang.org/x/crypto v0.7.0
	golang.org/x/net v0.8.0
	gopkg.in/couchbase/gocb.v1 v1.6.7
)

require (
	github.com/RoaringBitmap/roaring v1.2.3 // indirect
	github.com/aws/aws-sdk-go v1.44.170 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.2.2 // indirect
	github.com/blevesearch/bleve-mapping-ui v0.5.1 // indirect
	github.com/blevesearch/bleve/v2 v2.3.9-0.20230615140129-63e7a2fa1613 // indirect
	github.com/blevesearch/bleve_index_api v1.0.5 // indirect
	github.com/blevesearch/geo v0.1.17 // indirect
	github.com/blevesearch/go-porterstemmer v1.0.3 // indirect
	github.com/blevesearch/goleveldb v1.0.1 // indirect
	github.com/blevesearch/gtreap v0.1.1 // indirect
	github.com/blevesearch/mmap-go v1.0.4 // indirect
	github.com/blevesearch/scorch_segment_api/v2 v2.1.5 // indirect
	github.com/blevesearch/sear v0.1.0 // indirect
	github.com/blevesearch/segment v0.9.1 // indirect
	github.com/blevesearch/snowballstem v0.9.0 // indirect
	github.com/blevesearch/upsidedown_store_api v1.0.2 // indirect
	github.com/blevesearch/vellum v1.0.9 // indirect
	github.com/blevesearch/zapx/v11 v11.3.8 // indirect
	github.com/blevesearch/zapx/v12 v12.3.8 // indirect
	github.com/blevesearch/zapx/v13 v13.3.8 // indirect
	github.com/blevesearch/zapx/v14 v14.3.8 // indirect
	github.com/blevesearch/zapx/v15 v15.3.11 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cloudfoundry/gosigar v1.3.4 // indirect
	github.com/couchbase/blance v0.1.3 // indirect
	github.com/couchbase/cbft v0.0.0-00010101000000-000000000000 // indirect
	github.com/couchbase/cbgt v0.0.0-00010101000000-000000000000 // indirect
	github.com/couchbase/ghistogram v0.1.0 // indirect
	github.com/couchbase/gocb/v2 v2.5.4 // indirect
	github.com/couchbase/gocbcore/v9 v9.1.8 // indirect
	github.com/couchbase/gometa v0.0.0-20220803182802-05cb6b2e299f // indirect
	github.com/couchbase/hebrew v0.0.0-00010101000000-000000000000 // indirect
	github.com/couchbase/moss v0.3.0 // indirect
	github.com/couchbase/tools-common v0.0.0-20230525144302-671fb9dd857e // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/go-jsonpointer v0.0.0-20140810065344-75939f54b39e // indirect
	github.com/dustin/gojson v0.0.0-20150115165335-af16e0e771e2 // indirect
	github.com/elazarl/go-bindata-assetfs v1.0.1 // indirect
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/flatbuffers v22.11.23+incompatible // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.13.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/santhosh-tekuri/jsonschema v1.2.4 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/stretchr/testify v1.8.2 // indirect
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	go.etcd.io/bbolt v1.3.7 // indirect
	golang.org/x/exp v0.0.0-20221229233502-02c3fc3b3eb4 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/term v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	google.golang.org/genproto v0.0.0-20221227171554-f9683d7f8bef // indirect
	google.golang.org/grpc v1.51.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/couchbase/gocbcore.v7 v7.1.18 // indirect
	gopkg.in/couchbaselabs/gocbconnstr.v1 v1.0.4 // indirect
	gopkg.in/couchbaselabs/gojcbmock.v1 v1.0.4 // indirect
	gopkg.in/couchbaselabs/jsonx.v1 v1.0.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
