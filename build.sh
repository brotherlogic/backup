protoc --proto_path ../../../ -I=./proto --go_out=plugins=grpc:./proto proto/backup.proto
mv proto/github.com/brotherlogic/backup/proto/* ./proto
