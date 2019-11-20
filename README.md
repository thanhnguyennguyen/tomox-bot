# tomox-bot
send orders to tomox rpc

## Build
```
go get github.com/tomochain/tomochain
go build bot.go
```

## Run
 ```
 ./bot $BOT_ADDR $BOT_ADDR_KEY 6 1000
# send tomox orders starting with orderNonce = 6, speed : 1 order/second
 ```
 Params: 
 - Public address of bot
 - Private key of bot
 - StartNonce
 - BreakTime between 2 orders
 
