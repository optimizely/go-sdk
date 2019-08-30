# Optimizely Go SDK Examples


### Content
There are 2 functions:

* `examples()`
* `stressTest()`

`examples()` shows how to instantiate and execute various versions of Optimizely Clients

`stressTest()` is used to stress test our optimizely app and run cpu and memory profiling


### Profiling
Prerequisite: `go get github.com/pkg/profile`

* CPU profile. Execute: `go build -ldflags "-X main.RunCPUProfile=true" main.go && ./main`. It will create cpu.pprof file in your current directory. Then run: `go tool pprof -http=:8080 cpu.pprof` and profile cpu usage using web browser.
* Memory profile. Execute: `go build -ldflags "-X main.RunMemProfile=true" main.go && ./main`. It will create mem.pprof file in your current directory. Then run: `go tool pprof -http=:8080 mem.pprof` and profile memory using browser.
