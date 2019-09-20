# tomox-bot
send orders to tomox rpc

This bot helps to increase liquidity for the market, not for taking profit

## Build

- Build binary for your current platform 
```bash
go build
```
- Build binaries for multiple platforms<br/> 
_(Not recommended: because it takes a few minutes)_

  - Docker installed is required
  - Update TOMOX_BOT_PACKAGE, TOMOX_BOT_VERSION in build.sh, point to your remote package
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
