export TOMOX_BOT_PACKAGE=github.com/thanhnguyennguyen/tomox-bot
export TOMOX_BOT_VERSION=0.8.1
export GO111MODULE=off
echo Deleting old binaries in 'bin'
sudo rm -rf bin
git add build.sh
git commit -m "bump$TOMOX_BOT_VERSION"
git push origin master

gittool -t v$TOMOX_BOT_VERSION v$TOMOX_BOT_VERSION


start=`date +%s`

go get -v github.com/karalabe/xgo # Go CGO cross compiler
go build $GOPATH/src/github.com/karalabe/xgo/xgo.go


./xgo -go 1.12.5 -targets="linux/amd64,linux/386,darwin/amd64,darwin/386,windows/amd64,windows/386" -out bin/tomox-bot-$TOMOX_BOT_VERSION $TOMOX_BOT_PACKAGE

gittool -u v$TOMOX_BOT_VERSION tomox-bot-$TOMOX_BOT_VERSION-windows-4.0-amd64.exe
gittool -u v$TOMOX_BOT_VERSION tomox-bot-$TOMOX_BOT_VERSION-windows-4.0-386.exe
gittool -u v$TOMOX_BOT_VERSION tomox-bot-$TOMOX_BOT_VERSION-linux-amd64
gittool -u v$TOMOX_BOT_VERSION tomox-bot-$TOMOX_BOT_VERSION-linux-386
gittool -u v$TOMOX_BOT_VERSION tomox-bot-$TOMOX_BOT_VERSION-darwin-10.6-amd64
gittool -u v$TOMOX_BOT_VERSION tomox-bot-$TOMOX_BOT_VERSION-darwin-10.6-386


echo It took $((($(date +%s)-$start)/60)) minutes

