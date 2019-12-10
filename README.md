# tomox-bot
send orders to tomox rpc

This bot helps to increase liquidity for the market, not for taking profit

## Build

- 
```bash
go build -v -o tomox-bot
```

## Run
 ```bash
 cp .env.sample .env 
 # update .env file 

 
 # run bot 
 ./bot
 ```

for `cancel order`
```bash
    ./bot cancel [orderId]
```
We assume the baseToken, quoteToken of cancel orders defined in .env file, same pair as `new order`
 
## Download binaries
[https://github.com/thanhnguyennguyen/tomox-bot/releases](https://github.com/thanhnguyennguyen/tomox-bot/releases)
