# Project Rules

## 后端开发流程

- 修改后端代码后必须重新编译并重启服务
- 编译命令: `cd api-server && go build -o bin/api-server ./cmd/api-server`
- 启动命令: `nohup ./bin/api-server -config configs/api-server.yaml > /tmp/api-server.log 2>&1 &`
- 重启前先 `kill` 旧进程，再启动新进程

## task_monitor server 开发流程

- 代码位置: `/root/task_monitor/task_monitor_go/`
- 编译命令: `cd /root/task_monitor/task_monitor_go && go build -o bin/server ./cmd/server`
- 启动命令: `nohup ./bin/server -config configs/server.yaml > /tmp/task_monitor_server.log 2>&1 &`
- 重启前先 `kill` 旧进程，再启动新进程

## Agent 编译与打包

- 代码位置: `/root/task_monitor/task_monitor_go/`
- 本机编译: `cd /root/task_monitor/task_monitor_go && go build -o bin/agent ./cmd/agent`
- ARM64 交叉编译: `GOOS=linux GOARCH=arm64 go build -o bin/agent_arm64 ./cmd/agent`
- 打包部署包（参考 `task_monitor_agent_arm64_deploy.tar.gz` 格式）:
  ```bash
  cd /root/task_monitor/task_monitor_go
  mkdir -p task_monitor_agent_deploy
  cp bin/agent task_monitor_agent_deploy/
  cp configs/agent.yaml task_monitor_agent_deploy/
  cp scripts/deploy_server.sh task_monitor_agent_deploy/install.sh 2>/dev/null || true
  tar czf task_monitor_agent_deploy.tar.gz task_monitor_agent_deploy/
  rm -rf task_monitor_agent_deploy
  ```
- 部署包内容: `agent`(二进制) + `agent.yaml`(配置) + `install.sh`(安装脚本) + `uninstall_agent.sh`(卸载脚本)
- 目标机器安装目录: `/opt/task_monitor/`，通过 systemd 管理服务
