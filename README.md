# tomox-bot
send orders to tomox rpc

This bot helps to increase liquidity for the market, not for taking profit

## Build
```
./build.sh
```

## Run
 ```bash
 cp .env.sample .env 
 # update .env file 
 
 # see all binaries at /build, copy binary corresponding your os
 cp build/tomox-bot-linux-amd64 ./bot
 
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
