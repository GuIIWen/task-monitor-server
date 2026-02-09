# Project Rules

## 后端开发流程

- 修改后端代码后必须重新编译并重启服务
- 编译命令: `cd api-server && go build -o bin/api-server ./cmd/api-server`
- 启动命令: `nohup ./bin/api-server -config configs/api-server.yaml > /tmp/api-server.log 2>&1 &`
- 重启前先 `kill` 旧进程，再启动新进程
