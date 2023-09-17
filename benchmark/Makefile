all: yamly_bench_yayamls_engine yamly_bench_go_yaml_engine goccy_bench go_yaml_bench

yamly_bench_yayamls_engine:
	go run ../cmd/yamlygen/main.go -build-tags bench_yamly_yayamls_engine -engine yayamls -type LargeStruct
	go run ../cmd/yamlygen/main.go -build-tags bench_yamly_yayamls_engine -engine yayamls -type ExtraLargeStruct
	go run ../cmd/yamlygen/main.go -build-tags bench_yamly_yayamls_engine -engine yayamls -type SmallStruct

	#go test -cpuprofile cpu.out -memprofilerate=1 -memprofile mem.out -benchmem -tags bench_yamly -bench .
	go test -benchmem -tags bench_yamly_yayamls_engine -bench .
	rm *_yamly.go

yamly_bench_go_yaml_engine:
	go run ../cmd/yamlygen/main.go -build-tags bench_yamly_go_yaml_engine -engine goyaml -type LargeStruct
	go run ../cmd/yamlygen/main.go -build-tags bench_yamly_go_yaml_engine -engine goyaml -type ExtraLargeStruct
	go run ../cmd/yamlygen/main.go -build-tags bench_yamly_go_yaml_engine -engine goyaml -type SmallStruct

	go test -benchmem -tags bench_yamly_go_yaml_engine -bench .
	rm *_yamly.go

goccy_bench:
	go test -benchmem -tags bench_goccy -bench .

go_yaml_bench:
	go test -benchmem -tags bench_go_yaml -bench .
