# Serializer Benchmarks
Simple benchmark to see how some serializers I'm interested in work. This is an **unfair** test. 
It started with a zero alloc Cap'n Proto fork and added the others just to compare.

I will probably add more simple benchmarks to test features I'm interested in. Focusing on Cap'n Proto.

Cap'n Proto, Bebop, and Protobuf.

| Lib         | Note                                                 | Package                              |
|-------------|------------------------------------------------------|--------------------------------------|
| Cap'n Proto | Writes seem annoying? Documentation could be better. | **FORKED** capnproto.org/go/capnp/v3 |
| bebop       | API seems off? / Limited                             | wellquite.org/bebop                  |
| Protobuf    | Lots of support. Slightly Slower and allocs.         | google.golang.org/protobuf           |


## Cap'n Proto Notes
Cap'n Proto out of box doesn't support zero allocs. So using [knervous' fork](https://github.com/knervous/go-capnp).


## Run Tests
```sh
go test -bench=.
```

## Generating files
Regenerate files (Might mess up on windows):
```sh
go generate ./...
```

## Results
```
goos: linux
goarch: amd64
pkg: github.com/TheGreatSage/capntest
cpu: AMD Ryzen 7 5800X 8-Core Processor             
BenchmarkNewMessage/No_Write-16         	   	 9583848	       125.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkNewMessage/Write-16            	    	 2136864	       560.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshal/Unmarshal-16         	    	 4582036	       261.6 ns/op	     352 B/op	       3 allocs/op
BenchmarkUnmarshal/UnmarshalZeroTo-16   	    	 8095225	       145.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshal/UnmarshalZeroThree-16         	 7764084	       153.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshal/Deserialize-16                	 4222326	       280.8 ns/op	     352 B/op	       3 allocs/op
BenchmarkUnmarshal/DeserializeZeroThree-16       	 7201585	       166.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshal/DeserializeZeroTo-16          	 7513872	       159.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshal/Beop-16                       	13293169	        87.94 ns/op	     144 B/op	       5 allocs/op
BenchmarkUnmarshal/DesccBeop-16                  	10009774	       120.4 ns/op	     240 B/op	       6 allocs/op
BenchmarkUnmarshal/Pro-16                        	 4402192	       272.9 ns/op	     288 B/op	       6 allocs/op
BenchmarkMarshal/Marshal-16                      	21089367	        48.40 ns/op	     208 B/op	       1 allocs/op
BenchmarkMarshal/MarshalTo-16                    	 1000000	      1132 ns/op	    9472 B/op	       1 allocs/op
BenchmarkMarshal/MarshalThree-16                 	92089093	        12.34 ns/op	       0 B/op	       0 allocs/op
BenchmarkMarshal/NewMarshalTo-16                 	67658379	        17.81 ns/op	       0 B/op	       0 allocs/op
BenchmarkMarshal/NewMarshalThree-16              	95862961	        12.20 ns/op	       0 B/op	       0 allocs/op
BenchmarkMarshal/Beop-16                         	52134501	        22.50 ns/op	       0 B/op	       0 allocs/op
BenchmarkMarshal/Pro-16                          	 6938816	       172.7 ns/op	     128 B/op	       1 allocs/op
PASS
ok  	github.com/TheGreatSage/capntest	21.165s
```

