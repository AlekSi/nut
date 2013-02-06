all: integration_test_short

prepare:
	go env
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

integration_test_short: test
	go test -v -short github.com/AlekSi/nut/integration_test -gocheck.v

integration_test: test
	GONUTS_IO_SERVER=localhost:8080 go test -v github.com/AlekSi/nut/integration_test -gocheck.v
