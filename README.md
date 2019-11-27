# tomox-bot
send orders to tomox rpc

## Build
```
./build.sh
```

## Run
 ```
 cp .env.sample .env 
 # update .env file 
 
 # see all binaries at /build, copy binary corresponding your os
 cp build/tomox-bot-linux-amd64 ./bot
 
 # run bot 
 ./bot 6
# send tomox orders starting with orderNonce = 6
 ```
 Argument: 
 - StartNonce

 
