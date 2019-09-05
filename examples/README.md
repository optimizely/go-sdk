# Optimizely Go SDK Examples

Examples of client instantiation, datafile management and Feature enabled are provided in main.go 

# Profiling

Prerequisite:
1. `go get github.com/pkg/profile`
2. DATAFILES_DIR env var has to be set to point to the path where 100_entities.json is located

* CPU profile. Execute: `go build -ldflags "-X main.ProfileMode=cpu" main_profile_feature.go && ./main_profile_feature`. It will create cpu.pprof file in your current directory. Then run: `go tool pprof -http=:8080 cpu.pprof` and profile cpu usage using web browser.
* Memory profile. Execute: `go build -ldflags "-X main.ProfileMode=mem" main_profile_feature.go.go && ./main_profile_feature`. It will create mem.pprof file in your current directory. Then run: `go tool pprof -http=:8080 mem.pprof` and profile memory using browser.
