all: yayamls_bench goccy_bench

yayamls_bench:
	go run ../cmd/yayamls/main.go -build-tags bench_yayamls -type LargeStruct
	go run ../cmd/yayamls/main.go -build-tags bench_yayamls -type ExtraLargeStruct
	go run ../cmd/yayamls/main.go -build-tags bench_yayamls -type SmallStruct
	go test -benchmem -tags bench_yayamls -bench .
	rm *_yayamls.go
	
goccy_bench:
	go test -benchmem -tags bench_goccy -bench .
