# grpc-intro-go

after instaling grpc package these should be run
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin


# for sending message by grpcurl for streaming requests:
grpcurl --plaintext --msg-template -d @ 127.0.0.1:9094 Currency.SubscribeRates