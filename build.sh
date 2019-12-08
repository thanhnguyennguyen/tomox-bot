export GO111MODULE=off
rm -rf bin
go get -v
for GOOS in darwin linux windows; do
    for GOARCH in 386 amd64; do
        env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=1 go build -v -o bin/tomox-bot-$GOOS-$GOARCH
    done
done

