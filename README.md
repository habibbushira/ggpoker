# gRPC setup
1. sudo apt install -y protobuf-compiler
2. go get -u google.golang.org/protobuf/cmd/protoc-gen-go-grpc
3. go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
4. make proto


## bash source
vi ~/.bash_profile
export GO_PATH=~/go
export PATH=$PATH:/$GO_PATH/bin
source ~/.bash_profile
