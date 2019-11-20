# tomox-bot
send orders to tomox protocol

## Build
go get github.com/tomochain/tomochain
go build bot.go

## Run
 ```
 ./bot $BOT_ADDR $BOT_ADDR_KEY 6 1000
 ```
 Params: 
 - Public address of bot
 - Private key of bot
 - StartNonce
 - BreakTime between 2 orders
 
