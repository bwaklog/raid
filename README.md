_Run the benchmark_

```sh
go test -bench=.
```

_Sample_ with MAX_TEST_WRITERS set to 1

```text
$ go test -bench=.
goos: darwin
goarch: arm64
pkg: github.com/bwaklog/raid
cpu: Apple M1
BenchmarkWALMode-8      	       6	 170083514 ns/op
BenchmarkDeleteMode-8   	       1	2256889750 ns/op
PASS
ok  	github.com/bwaklog/raid	4.388s
```

### Todo

- [ ] Retry for db lock errors
