export GO111MODULE=off
rm -rf build
go get -v
for GOOS in darwin linux windows; do
    for GOARCH in 386 amd64; do
        go build -v -o build/tomox-bot-$GOOS-$GOARCH
    done
done

