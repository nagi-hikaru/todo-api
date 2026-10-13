# Todo API

基于 Go、Gin 和 SQLite 的 REST 风格待办事项 API，支持任务增删改查、状态管理、关键词搜索和分页筛选，集成 Swagger 在线文档。

## 快速启动

环境要求：Go 1.26.2 或更高版本，以及可用的 C 编译器（SQLite 驱动依赖 CGO，需启用 `CGO_ENABLED=1`）。

在项目根目录执行：

```bash
go mod download
go run ./cmd/server
```

服务监听 `8080` 端口，首次启动会在当前工作目录创建 `todos.db` 和任务表，无需单独部署数据库。

- API 地址：[http://localhost:8080/api/v1/todos](http://localhost:8080/api/v1/todos)
- Swagger 文档：[http://localhost:8080/api/v1/docs/index.html](http://localhost:8080/api/v1/docs/index.html)

## 接口列表

所有接口以 `/api/v1` 为前缀。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/todos` | 分页查询任务，支持筛选和搜索 |
| GET | `/todos/{id}` | 查询任务详情 |
| POST | `/todos` | 创建任务 |
| PUT | `/todos/{id}` | 更新任务内容 |
| PATCH | `/todos/{id}/status?status=completed` | 更新任务状态 |
| DELETE | `/todos/{id}` | 删除任务 |

- 状态：`pending`、`completed`、`cancelled`；新任务默认为 `pending`。
- 优先级：`low`、`medium`、`high`，可不填写。
- 截止日期：`YYYY-MM-DD`，可不填写。
- 列表参数：`status`、`priority`、`search`、`page`、`page_size`，均可省略。
- 分页默认第 `1` 页、每页 `20` 条；页码至少为 `1`，每页条数为 `1～100`。
- PUT 请求必须提供标题；省略描述、优先级或截止日期会清空对应字段，状态保持不变。

## 请求示例

创建任务：

```bash
curl -X POST http://localhost:8080/api/v1/todos \
  -H 'Content-Type: application/json' \
  -d '{"title":"学习 Gin API","priority":"high","due_date":"2026-12-31"}'
```

筛选并分页查询：

```bash
curl 'http://localhost:8080/api/v1/todos?status=pending&search=Gin&page=1&page_size=10'
```

标记任务完成，将 `1` 替换为创建响应中 `data.id` 的实际值：

```bash
curl -X PATCH 'http://localhost:8080/api/v1/todos/1/status?status=completed'
```

更多请求见 [test.http](test.http)。可使用 VS Code 的 REST Client 扩展逐条发送请求；执行前设置实际的 `todoId`。其中的预期结果注释用于手动核对，不是自动化断言。

## 项目结构

```text
cmd/server/          服务入口
internal/models/     数据模型与请求、响应结构
internal/database/   SQLite 初始化和数据操作
internal/handler/    请求处理与参数校验
internal/middleware/ 日志、异常恢复和跨域配置
internal/router/     API 与 Swagger 路由
docs/                生成的 Swagger 文档
test.http            接口手动测试请求
```

## 更新 Swagger 文档

修改接口注释或模型后，在项目根目录重新生成文档，并重启服务：

```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/server/main.go --parseInternal -o docs
```

生成的 `docs/` 随代码提交，本地数据库 `todos.db` 已通过 `.gitignore` 忽略。
