module ermes-central-repo

go 1.24.3

require (
	github.com/ermes-thesis/ermes-wasm-pool v0.0.0-00010101000000-000000000000
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/google/uuid v1.6.0
	github.com/hashicorp/golang-lru/v2 v2.0.7
	github.com/redis/go-redis/v9 v9.9.0
	github.com/spf13/viper v1.21.0
	github.com/stretchr/testify v1.11.1
	golang.org/x/sync v0.16.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/second-state/WasmEdge-go v0.14.0 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/ermes-thesis/ermes-wasm-pool => ../ermes-wasm-pool
