# WCA 登录 - 前端对接手册

本文档对应 [`src/api/app/auth/wca.go`](../src/api/app/auth/wca.go) 与 [`src/api/routes/auth.go`](../src/api/routes/auth.go) 的现有实现。完整 API 前缀为 `/v3/cube-api`，配置结构位于 [`src/configs/global.go`](../src/configs/global.go)。

---

## 一、登录方式概览

| 登录方式 | 接口 | 说明 |
|---------|------|------|
| 账号密码登录 | `POST /v3/cube-api/auth/login` | 原有方式，保持不变 |
| **WCA 登录** | `GET /v3/cube-api/auth/wca` | 跳转到 WCA 授权页 |

两种方式最终都返回**相同格式的 JWT**，后续请求携带方式一致。

---

## 二、WCA 登录流程

```
┌─────────┐    ① 点击 WCA 登录     ┌─────────┐    ② 307 跳转     ┌─────────┐
│  前端   │ ────────────────────► │  后端   │ ───────────────► │  WCA    │
│         │   GET /auth/wca       │         │  授权页 URL      │  官网   │
└─────────┘   ?redirect=xxx       └─────────┘                  └────┬────┘
     ▲                                                                    │
     │                                                                    │ ③ 用户授权
     │                                                                    ▼
     │    ⑤ 302 跳回 + token      ┌─────────┐    ④ 回调 + code    ┌─────────┐
     └────────────────────────── │  后端   │ ◄───────────────── │  WCA    │
        或 ⑤' JSON 返回           │ callback│   /auth/wca/callback│         │
                                 └─────────┘                     └─────────┘
```

---

## 三、接口说明

### 1. 发起 WCA 登录

**请求**

```
GET /v3/cube-api/auth/wca?redirect={redirect}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| redirect | string | 否 | 登录成功后跳转地址，需符合 `global.wcaAuth2` 的允许列表，规则见下文 |

**行为**

- 307 重定向到 WCA 授权页
- `redirect` 为空或 `debug`：回调时返回 JSON，不跳转（调试用）
- `redirect` 为有效前端地址：回调时 302 跳转到 `redirect?token=xxx`

**redirect 校验**

- `global.wcaAuth2.redirectURLs` 长度大于 0 时，**只使用该列表**，不会再合并 `frontendBase` 或 `redirectBase`。列表非空但只含空字符串时也不会回退。
- 仅当 `redirectURLs` 为空时，使用非空 `frontendBase`，再加入与它不同且非空的 `redirectBase`。`global.baseHost` 不参与这个允许列表的回退。
- 当前代码去掉允许项末尾的一个 `/`，接受 `redirect` 的 `scheme://host` 与该项相等，或整个 `redirect` 以该项加 `/` 开头的情况；不是任意子域名通配。建议配置完整 origin（含协议和必要端口），不要依赖路径项做严格的路径隔离。
- 例如允许项 `http://localhost:3000` 可接受 `http://localhost:3000/profile`；`http://localhost:3001` 需单独配置。`localhost` 和 `127.0.0.1` 也应视为不同地址。
- 空 `redirect` 或字面值 `debug` 是调试特例，回调直接返回 JSON。

**错误响应**（JSON，未跳转时）

| http_code | error / data | 说明 |
|-----------|---------|------|
| 400 | invalid redirect URI | redirect 不在允许域名下 |
| 500 | WCA OAuth 未配置 | 后端未配置 appId/appSecret |

---

### 2. WCA 回调（后端处理，前端不直接调用）

```
GET /v3/cube-api/auth/wca/callback?code=xxx&state=xxx
```

由 WCA 授权后自动跳转，前端只需在合适的页面接收带 `token` 的跳转。

---

### 3. 调试：获取当前用户

**请求**

```
GET /v3/cube-api/auth/wca/me
Authorization: Bearer {token}
```

**响应**

```json
{
  "user": {
    "id": 1,
    "Name": "张三",
    "WcaID": "2020ZHAN01",
    "WcaLoginAt": "2026-03-11T12:00:00Z"
  }
}
```

---

## 四、前端需要实现的逻辑

### 1. WCA 登录入口

```javascript
// 跳转到后端 WCA 登录接口，redirect 为登录成功后要回到的页面
const redirect = encodeURIComponent(`${window.location.origin}/profile`);
window.location.href = `${API_BASE}/v3/cube-api/auth/wca?redirect=${redirect}`;
```

### 2. 接收 token 的页面

登录成功后会 302 到 `redirect?token=xxx`，需在对应页面（如 `/profile`、`/auth/callback`）处理：

```javascript
// 在页面加载时检查 URL 是否带有 token
const params = new URLSearchParams(window.location.search);
const token = params.get('token');
if (token) {
  // 1. 存储 token（与账号密码登录一致）
  localStorage.setItem('token', token);
  // 2. 清除 URL 中的 token，避免泄露
  const url = new URL(window.location.href);
  url.searchParams.delete('token');
  window.history.replaceState({}, '', url.toString());
  // 3. 更新登录状态，跳转到目标页等
}
```

### 3. 统一 token 使用方式

WCA 登录与账号密码登录的 JWT 格式相同，请求方式一致：

```
Authorization: Bearer {token}
```

当前中间件从 `Authorization` 请求头读取 token（`TokenLookup: "header: Authorization"`）。不要因为账号密码登录配置了 `SendCookie` 就假定 WCA 回调会设置登录 Cookie，或受保护接口会从 Cookie 读取 token；WCA 回调使用 `TokenGenerator` 后跳转/返回 JSON。

---

## 五、数据结构变更

### 1. 用户 claims（与原有一致）

以下是 `JwtMapClaims` 业务字段；JWT 中由中间件放在 `UserDetail` 下，不是 token 顶层的完整结构。调试回调 JSON 的 `user` 则直接使用该结构。

```typescript
interface JwtMapClaims {
  id: number;
  auth: number;        // 权限位
  name: string;
  enName: string;
  loginID: string;
  cubeID: string;
  wcaID: string;       // WCA 登录用户会有值
  delegateName: string;
}
```

### 2. 用户信息（GET /auth/current 等）

**新增字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| WcaID | string | WCA ID，如 "2020ZHAN01" |
| WcaLoginAt | string (ISO 8601) | 最近一次 WCA 登录时间；模型使用 `omitempty`，未设置时字段可省略 |

**不返回字段**（仅后端存储）

- `WcaAccessToken`
- `WcaTokenExpiresAt`

### 3. 调试模式回调（redirect=debug 或空）

**响应**（JSON，不跳转）

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "auth": 1,
    "name": "张三",
    "enName": "Zhang San",
    "loginID": "",
    "cubeID": "",
    "wcaID": "2020ZHAN01",
    "delegateName": ""
  },
  "wca": {
    "wca_id": "2020ZHAN01",
    "name": "张三",
    "email": "user@example.com",
    "avatar": "https://..."
  }
}
```

---

## 六、错误响应格式

与现有接口一致：

```json
{
  "http_code": 400,
  "message": "无效输入",
  "error": "invalid redirect URI",
  "data": "invalid redirect URI",
  "level": "",
  "line": "err at ...",
  "ref": ""
}
```

---

## 七、注意事项

### 1. redirect 安全

- `redirect` 必须符合上文的允许列表规则；`redirectURLs` 非空时优先且独占
- 不要用用户输入拼接 redirect，避免开放重定向
- 成功跳转会在 URL 查询参数携带 JWT；接收页面应及时清除 token，并避免将完整回调 URL 写入日志或分享

### 2. 环境与回调地址

| 环境示例 | 回调地址 |
|------|----------|
| 生产 | `https://cubing.pro/v3/cube-api/auth/wca/callback` |
| 本地 Gateway（公开示例端口） | `http://127.0.0.1:8080/v3/cube-api/auth/wca/callback` |
| 直连 API | `http://localhost:20000/v3/cube-api/auth/wca/callback` |

端口仅为示例，必须与实际监听/代理地址及 WCA 应用登记值一致。后端优先用 `global.wcaAuth2.redirectBase`，为空时用 `global.baseHost`，去掉末尾 `/` 后拼接 `/v3/cube-api/auth/wca/callback`。`redirectURLs` 和 `frontendBase` 用于前端返回地址校验，不决定这个 OAuth 回调地址。

前端发起登录时应使用当前环境的 API 根地址；上文 JavaScript 的 `API_BASE` 不包含 `/v3/cube-api` 前缀，避免重复拼接。

### 3. 自动重新授权

以下情况后端会尝试创建新的 state 并以 **307** 重新跳转 WCA 授权页：

- state 解码成功，但数据库中找不到对应 nonce 或记录已过期
- 使用 code 换取 token 失败（不限于过期）

缺失 code/state、state 编码或 JSON 格式错误等直接返回错误，不会自动重试。重新授权也依赖数据库和 WCA 服务，不保证一定成功。

### 4. WCA 用户与本地用户

- 首次 WCA 登录：创建新用户，`WcaID` 必填
- 再次 WCA 登录：若 `WcaID` 已存在则关联并更新信息
- 账号密码登录与 WCA 登录共用同一套用户体系，通过 `WcaID` 关联

### 5. 调试建议

- 使用 `?redirect=debug` 可在回调时直接拿到 JSON，便于调试
- `/auth/wca/me` 需携带有效 token，用于验证登录状态

---

## 八、接口汇总

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| GET | /v3/cube-api/auth/wca | 否 | 发起 WCA 登录 |
| GET | /v3/cube-api/auth/wca/callback | 否 | WCA 回调（后端处理） |
| GET | /v3/cube-api/auth/wca/me | Bearer | 调试：获取当前用户 |
| POST | /v3/cube-api/auth/login | 否 | 账号密码登录（原有） |
| GET | /v3/cube-api/auth/current | Bearer | 获取当前用户（原有） |

---

## 九、配置参考（后端）

```yaml
# 填入被忽略的本地配置；凭据需自行替换，不提交仓库
global:
  baseHost: "http://localhost:20000"
  wcaAuth2:
    appId: "YOUR_WCA_APP_ID"
    appSecret: "YOUR_WCA_APP_SECRET"
    redirectBase: "http://localhost:20000" # 后端 OAuth 回调根地址
    frontendBase: "http://localhost:3000" # 仅在 redirectURLs 为空时参与允许列表
    redirectURLs:                       # 非空时只允许此列表
      - "http://localhost:3000"
    auths: ["public", "dob", "email", "openid", "profile"]
```

这是配置片段，需合并到已有 `global` 下，不能覆盖其数据库等其他字段。本例对应的 WCA 应用回调地址为 `http://localhost:20000/v3/cube-api/auth/wca/callback`；通过 Gateway 访问时按实际地址修改。`auths` 为空时当前代码使用相同的默认 scopes。

配置加载不读取旧示例中的 `WCA_CLIENT_ID` 等环境变量。详细说明见[配置文档](configuration.md)，旧 MySQL 库的补迁移条件见[脚本说明](../scripts/README.md)。

---

*文档按仓库当前实现核对；不代表已完成 WCA 在线联调。*
