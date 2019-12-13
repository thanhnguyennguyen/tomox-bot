export GO111MODULE=off
echo Deleting old binaries in 'bin'
sudo rm -rf bin
go get -v gopkg.in/natefinch/npipe.v2
go get -v github.com/karalabe/xgo # Go CGO cross compiler
go build $GOPATH/src/github.com/karalabe/xgo/xgo.go


./xgo -go 1.12.5 -targets="linux/amd64,linux/386,darwin/amd64,darwin/386,windows/amd64,windows/386" -out bin/tomox-bot-v0.7 github.com/thanhnguyennguyen/tomox-bot

