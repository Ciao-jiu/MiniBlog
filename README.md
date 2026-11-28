# MiniBlog

一个使用 Go 实现的文章管理 API，包含用户注册登录、JWT 认证、文章 CRUD、作者权限校验、分页查询和 Redis 文章列表缓存。

## 技术栈

- Go、Gin
- GORM、MySQL
- Redis
- JWT、bcrypt
- Vue 3、Vite（前端位于 `web/`）

## 功能

- 用户注册和登录，密码使用 bcrypt 哈希存储。
- 使用 JWT 保护创建、更新和删除文章接口。
- 文章作者校验，禁止非作者修改或删除文章。
- MySQL 持久化用户和文章数据，支持文章详情和分页列表查询。
- Redis 缓存分页文章列表；缓存读取或写入失败时记录日志，列表读取回退到 MySQL。
- 文章写操作成功后清理文章列表缓存。

## 环境要求

- Go 版本以 [`go.mod`](go.mod) 中的声明为准。
- MySQL
- Redis

## 本地启动

1. 启动 MySQL 和 Redis。
2. 在 MySQL 中创建开发数据库：

   ```sql
   CREATE DATABASE miniblog CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

3. 根据 `.env.example` 中的示例，在当前终端设置环境变量。项目目前通过 `os.Getenv` 读取系统环境变量，**不会自动加载 `.env` 文件**。

   PowerShell 示例（请替换 MySQL 用户名和密码）：

   ```powershell
   $env:MINIBLOG_DB_DSN = 'root:你的MySQL密码@tcp(127.0.0.1:3306)/miniblog?charset=utf8mb4&parseTime=True&loc=Local'
   $env:MINIBLOG_REDIS_ADDR = '127.0.0.1:6379'
   $env:MINIBLOG_JWT_SECRET = '替换为你自己的随机长字符串'
   ```

4. 在项目根目录启动：

   ```powershell
   go run .
   ```

服务默认监听 `:8080`。启动时会连接 MySQL、自动迁移用户和文章表，并检查 Redis 连接。上述环境变量只在当前 PowerShell 窗口有效。

## API

| 方法 | 路径 | 说明 | 认证 |
| --- | --- | --- | --- |
| `POST` | `/api/v1/register` | 注册用户 | 否 |
| `POST` | `/api/v1/login` | 登录并获取 JWT | 否 |
| `GET` | `/api/v1/articles` | 分页查询文章 | 否 |
| `GET` | `/api/v1/articles/:id` | 查询文章详情 | 否 |
| `POST` | `/api/v1/articles` | 创建文章 | 是 |
| `PUT` | `/api/v1/articles/:id` | 更新文章（仅作者） | 是 |
| `DELETE` | `/api/v1/articles/:id` | 删除文章（仅作者） | 是 |

受保护接口在请求头中携带登录接口返回的 Token：

```text
Authorization: Bearer <JWT>
```

文章列表支持 `page` 和 `page_size` 查询参数，例如：

```text
GET /api/v1/articles?page=1&page_size=10
```

注册请求示例：

```json
{"username":"alice","password":"change-this-password"}
```

## 测试

集成测试需要 MySQL 和 Redis。先创建**独立测试数据库**：

```sql
CREATE DATABASE miniblog_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

然后在 PowerShell 设置测试环境变量（替换 MySQL 密码）：

```powershell
$env:MINIBLOG_TEST_DB_DSN = 'root:你的MySQL密码@tcp(127.0.0.1:3306)/miniblog_test?charset=utf8mb4&parseTime=True&loc=Local'
$env:MINIBLOG_TEST_REDIS_ADDR = '127.0.0.1:6379'
go test ./...
```

**数据安全提醒：** 测试初始化会迁移并清理测试数据库中的 `users`、`articles` 表，还会对 `MINIBLOG_TEST_REDIS_ADDR` 指向的 Redis 执行 `FLUSHDB`。必须使用专用测试数据库和可清空的 Redis 实例/数据库，不要指向开发或生产数据。测试 Redis 默认使用本机地址时也会清空 Redis 0 号库。

静态检查：

```powershell
go vet ./...
```

## 前端

前端在 `web/`，依赖安装和开发服务器启动方法见 [`web/README.md`](web/README.md)。

## 项目范围

本项目用于学习和演示基础后端开发流程，目前没有性能压测数据，不宣称高并发或高可用能力。生产使用前还需要补充安全配置、服务优雅关闭、连接池与超时治理等工作。
