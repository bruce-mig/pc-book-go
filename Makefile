gen:
	rm -f pb/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	proto/*.proto \
	--grpc-gateway_out=:pb --openapiv2_out=:swagger

server1:
	go run cmd/server/main.go -port 50051

server2:
	go run cmd/server/main.go -port 50052 

server1-tls:
	go run cmd/server/main.go -port 50051 -tls

server2-tls:
	go run cmd/server/main.go -port 50052 -tls

server:
	go run cmd/server/main.go -port 8080

rest:
	go run cmd/server/main.go -port 8081 -type rest -endpoint 0.0.0.0:8080

client:
	go run cmd/client/main.go -address 0.0.0.0:8080 

client-tls:
	go run cmd/client/main.go -address 0.0.0.0:8080 -tls

test:
	go test -cover -race ./...

cert:
	cd cert; ./gen.sh; cd ..

nginx:
	docker run --name pcbook-nginx-1.2 -v /home/migeri/Projects/pc-book/nginx/log/:/var/log/nginx/pcbook/ -d -p 8080:8080 pc-nginx:1.2

image:
	docker build -f nginx.dockerfile -t pc-nginx:1.2 .

.PHONY: gen server client test cert nginx image