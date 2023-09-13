all: yamly_bench goccy_bench go_yaml_bench

yamly_bench:
	go run ../cmd/yamlygen/main.go -build-tags bench_yamly -type LargeStruct
	go run ../cmd/yamlygen/main.go -build-tags bench_yamly -type ExtraLargeStruct
	go run ../cmd/yamlygen/main.go -build-tags bench_yamly -type SmallStruct
	#go test -cpuprofile cpu.out -memprofilerate=1 -memprofile mem.out -benchmem -tags bench_yamly -bench .
	go test -benchmem -tags bench_yamly -bench .
	rm *_yamly.go
	
goccy_bench:
	go test -benchmem -tags bench_goccy -bench .

go_yaml_bench:
	go test -benchmem -tags bench_go_yaml -bench .
