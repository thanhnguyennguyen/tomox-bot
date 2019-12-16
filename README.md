# tomox-bot
send orders to tomox rpc

This bot helps to increase liquidity for the market, not for taking profit

## Build

- Build binary for your current platform 
```bash
go build -v -o tomox-bot
```
- Build binaries for multiple platforms (Not recommendd: because it takes a few minutes)
    Docker installed is required
    Update TOMOX_BOT_PACKAGE in build.sh, point to your remote package
```bash
./build.sh
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
