all: test

prepare:
	go get -u launchpad.net/gocheck
	go get -u github.com/AlekSi/test_nut1
	-go get -u github.com/AlekSi/test_nut2
	-go get -u github.com/AlekSi/test_nut3

# format, vet, build
fvb:
	gofmt -e -s -w .
	go tool vet .
	go build -o gonut github.com/AlekSi/nut/nut

test: fvb
	cd ../test_nut1 && ../nut/gonut pack
	go test -v github.com/AlekSi/nut -gocheck.v
	go test -v github.com/AlekSi/nut/nut -gocheck.v
	go test -v github.com/AlekSi/nut/integration_test -gocheck.v

test_server: test
	cd ../test_nut1 && GONUTS_IO_SERVER=localhost:8080 ../nut/gonut publish -v test_nut1-0.0.1.nut
	cd ../test_nut2 && GONUTS_IO_SERVER=localhost:8080 ../nut/gonut publish -v test_nut2-0.0.2.nut
	cd ../test_nut1 && GONUTS_IO_SERVER=localhost:8080 ../nut/gonut get -v test_nut2/0.0.2
