all: yayamls_bench goccy_bench go_yaml_bench

yayamls_bench:
	go run ../cmd/yayamls/main.go -build-tags bench_yayamls -type LargeStruct
	go run ../cmd/yayamls/main.go -build-tags bench_yayamls -type ExtraLargeStruct
	go run ../cmd/yayamls/main.go -build-tags bench_yayamls -type SmallStruct
	go test -cpuprofile cpu.out -memprofilerate=1 -memprofile mem.out -benchmem -tags bench_yayamls -bench .
	#go test -benchmem -tags bench_yayamls -bench .
	rm *_yayamls.go
	
goccy_bench:
	go test -benchmem -tags bench_goccy -bench .

go_yaml_bench:
	go test -benchmem -tags bench_go_yaml -bench .
