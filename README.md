## evans実行方法
evans -r repl -p 50051 

package workout

service WorkoutService

show service

## proto生成コマンド
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/workout.proto