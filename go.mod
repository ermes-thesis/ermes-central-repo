module ermes-central-repo

go 1.24.3

require (
	github.com/ermes-thesis/ermes-common/caches v0.0.0-00010101000000-000000000000
	github.com/ermes-thesis/ermes-common/logger v0.0.0-00010101000000-000000000000
	github.com/ermes-thesis/ermes-common/model v0.0.0-00010101000000-000000000000
	github.com/ermes-thesis/ermes-common/wasm-pool v0.0.0-00010101000000-000000000000
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/google/uuid v1.6.0
	github.com/redis/go-redis/v9 v9.9.0
	github.com/rs/zerolog v1.34.0
	github.com/spf13/pflag v1.0.10
	github.com/spf13/viper v1.21.0
	github.com/stretchr/testify v1.11.1
	golang.org/x/sync v0.16.0
)

replace github.com/ermes-thesis/ermes-common/wasm-pool => ../ermes-common/wasm-pool

replace github.com/ermes-thesis/ermes-common/logger => ../ermes-common/logger

replace github.com/ermes-thesis/ermes-common/model => ../ermes-common/model

replace github.com/ermes-thesis/ermes-common/caches => ../ermes-common/caches

replace github.com/ermes-thesis/ermes-common/utils => ../ermes-common/utils

require (
	github.com/bytecodealliance/wasmtime-go/v37 v37.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/ermes-thesis/ermes-common/utils v0.0.0-00010101000000-000000000000 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/second-state/WasmEdge-go v0.14.0 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tetratelabs/wazero v1.9.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
