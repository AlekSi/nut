GO?=go

all: integration_test_short

prepare:
	$(GO) env
	$(GO) get -u launchpad.net/gocheck
	$(GO) get -u github.com/AlekSi/test_nut1
	-$(GO) get -u github.com/AlekSi/test_nut2
	-$(GO) get -u github.com/AlekSi/test_nut3

# format, vet, build
fvb:
	gofmt -e -s -w .
	$(GO) tool vet .
	$(GO) install github.com/AlekSi/nut
	$(GO) build -o gonut.exe github.com/AlekSi/nut/nut

test: fvb
	cd ../test_nut1 && ../nut/gonut.exe pack
	$(GO) test -v github.com/AlekSi/nut -gocheck.v
	$(GO) test -v github.com/AlekSi/nut/nut -gocheck.v

integration_test_short: test
	$(GO) test -v -short github.com/AlekSi/nut/integration_test -gocheck.v

integration_test: test
	GONUTS_IO_SERVER=localhost:8080 $(GO) test -v github.com/AlekSi/nut/integration_test -gocheck.v
