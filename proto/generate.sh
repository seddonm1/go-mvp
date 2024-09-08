# see https://github.com/speakeasy-api/speakeasy-grpc-gateway-example/tree/main
buf generate &&\
go run convert.go &&\
rm openapi/helloworld.swagger.json