# Fork 版统计数据持久化与上游同步指南

本文档用于你的 fork 项目：让管理面板的统计数据在重启后自动保留，并且尽量不影响后续跟进上游新版本。

## 1. 这次改动做了什么

- 新增自动持久化逻辑：服务启动时自动从 `auth-dir/usage-statistics.json` 恢复统计。
- 服务运行中每 30 秒自动落盘一次（有新请求时）。
- 服务退出时再做一次落盘，减少数据丢失。

默认文件位置基于 `auth-dir`：
- 本地运行（默认配置）：`~/.cli-proxy-api/usage-statistics.json`
- Docker Compose 默认映射：宿主机 `./auths/usage-statistics.json`

## 2. 你需要做的配置

确保统计功能开启（否则不会新增统计）：

```yaml
usage-statistics-enabled: true
```

推荐同时确认 `auth-dir` 落在可持久化目录（默认已是）：

```yaml
auth-dir: "~/.cli-proxy-api"
```

Docker 下请确保 `docker-compose.yml` 里有这条 volume（项目默认已有）：

```yaml
- ${CLI_PROXY_AUTH_PATH:-./auths}:/root/.cli-proxy-api
```

## 3. 如何验证是否生效

1. 启动服务并发送几次请求。
2. 查看统计接口有数据：`GET /v0/management/usage`
3. 停服务再启动。
4. 再次查看统计接口，确认数据仍在。
5. 检查本地文件是否存在：`auths/usage-statistics.json`（Docker）或 `~/.cli-proxy-api/usage-statistics.json`（本地）。

## 4. 推荐的 fork 同步策略（最少冲突）

目标：上游更新后，你只需要重放一个很小的补丁。

### 一次性准备

```bash
git remote add upstream https://github.com/router-for-me/CLIProxyAPI.git
git fetch upstream
```

把这次改动集中在一个独立分支（例如 `feature/usage-stats-persist`），并保持少量提交。

### 每次同步上游

```bash
git fetch upstream
git checkout main
git rebase upstream/main
git checkout -B my-main-with-persist main
git cherry-pick <你的持久化提交SHA>
```

如有冲突，只需处理这一个补丁的冲突；处理后运行：

```bash
go test ./internal/usage ./sdk/cliproxy
```

### 日常建议

- 不要把大量定制改动混在这个补丁里。
- 每次上游更新后先同步，再重放补丁，再跑测试。
- 若上游未来内置了同类能力，可直接删除该补丁并回归上游实现。
