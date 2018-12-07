# include $(GOPATH)/src/github.com/inoc603/go-make/Makefile

# GO_MAIN := ./cmd
# CGO := 1

GO111MODULE:=on

build: build/$(shell go env GOOS)_$(shell go env GOARCH)/sshd

build/%/sshd: *.go
	GOOS=$(firstword $(subst _, ,$*)) GOARCH=$(lastword $(subst _, ,$*)) \
		CGO_ENABLED=0 GO111MODULE=$(GO111MODULE) \
		go build -o $@ ./cmd

