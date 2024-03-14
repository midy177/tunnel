SET CGO_ENABLED=0 SET GOOS=linux SET GOARCH=amd64 go build -gcflags=-trimpath=$GOPATH -asmflags=-trimpath=$GOPATH -ldflags "-w -s"

# CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -gcflags=-trimpath=$GOPATH -asmflags=-trimpath=$GOPATH -ldflags "-w -s"


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go test -c

./utils.test -test.v -test.run TestName
