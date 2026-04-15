# 主办团队与比赛群组（管理端核心功能）

本文档描述 **超级管理员** 在管理端对「主办团队（`user.Organizers`）」与「比赛群组（`competition.CompetitionGroup`）」的 HTTP 接口、权限约束与字段约定，供前端与运维对接。

**基路径**：所有接口均在 ` /v3/cube-api/admin ` 下（与 `src/api/api.go` 中路由一致）。

**鉴权**：请求需携带有效 JWT；服务端通过 `middleware.CheckAuthMiddlewareFunc(user.AuthSuperAdmin)` 校验 **仅超级管理员** 可访问管理端路由（见 `src/api/routes/admin.go`）。

---

## 1. 业务规则摘要

| 规则 | 说明 |
|------|------|
| 管理权限 | 主办团队与比赛群组的创建/修改/删除、主办成员增删，仅超级管理员可操作。 |
| 群组数量上限 | 单个主办团队最多绑定 **3** 个比赛群组，常量：`user.MaxCompetitionGroupsPerOrganizer`（`src/internel/database/model/user/sponsor.go`）。 |
| 空主办团队 | 允许 `leader_cube_id` 为空、成员列表为空；便于先建团队再指派组长。 |
| 主办成员 | 通过「成员接口」向 `ass_org_users` 添加 CubeID；可选同时授予用户 `AuthOrganizers`（主办权限位）。 |
| 撤销主办权限 | 调用「修改用户权限」取消主办位时，若用户仍是 **任一团体的组长或成员**，接口会拒绝，需先从团队中移出（见下文 `POST /admin/users/update_auth`）。 |

**校验方式**：请求体 **不使用** Gin `binding` 标签做字段拦截；各 Handler 在代码内对必填项、枚举、外键（用户是否存在）、数量上限等做校验（与项目 `app_utils.BindAll` 约定一致）。

---

## 2. 权限位（User.Auth）

定义见 `src/internel/database/model/user/user.go`（`Auth` 为按位标志）：

| 常量 | 值（十进制） | 含义 |
|------|----------------|------|
| `AuthPlayer` | 1 | 选手 |
| `AuthOrganizers` | 2 | 主办 |
| `AuthDelegates` | 4 | 代表 |
| `AuthAdmin` | 8 | 管理员 |
| `AuthSuperAdmin` | 16 | 超级管理员 |

管理端「修改用户权限」接口使用 `set` + `auth` 组合对 **指定位** 进行置位或清除（见第 4 节）。

---

## 3. 数据模型与存储约定

### 3.1 主办团队 `user.Organizers`

主要字段：

| 字段 | 说明 |
|------|------|
| `name` | 团队名称，唯一。 |
| `introduction` | 介绍（如 Markdown）。 |
| `email` | 联系邮箱。 |
| `leaderId` | 组长 CubeID，可为空。 |
| `ass_org_users` | 成员 CubeID 列表，JSON 数组字符串（由服务端维护）。 |
| `status` | 状态，见 `OrganizersStatus` 枚举（`NotUse`、`Using`、`Applying` 等）。 |
| `leader_remark` | 组长备注。 |
| `admin_msg` | 管理员留言。 |

### 3.2 比赛群组 `competition.CompetitionGroup`

| 字段 | 说明 |
|------|------|
| `name` | 群组展示名称。 |
| `orgId` | 所属主办团队 ID（GORM 列名 `orgId`，对应结构体 `OrganizersID`）。 |
| `qq_groups` / `qq_group_uid` / `wechat_groups` | 多值字段：服务端以 **JSON 数组字符串** 落库；读接口兼容历史单行或逗号分隔（见 `competition.StringListToDB` / `StringListFromDB`）。 |

**API 层**：请求与响应中多值字段使用 JSON 数组 `string[]`；与数据库之间的序列化由服务端统一处理。

---

## 4. 用户权限：授予 / 撤销主办

**路径**：`POST /v3/cube-api/admin/users/update_auth`

**请求体**：

```json
{
  "cube_id": "用户 CubeID",
  "set": true,
  "auth": 2
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `cube_id` | string | 目标用户 CubeID。 |
| `set` | bool | `true` 为置位 `auth` 中的位，`false` 为清除。 |
| `auth` | number | 权限位掩码，通常传单项，如主办为 `2`（`AuthOrganizers`）。 |

**撤销主办（`set: false` 且清除 `AuthOrganizers`）**：若该用户仍是 **任一主办团队的组长**（`leaderId`）或 **成员**（`ass_org_users` 中含其 CubeID），返回 **422**（`ErrValidationFailed`），提示需先从主办团队中移出。

实现：`src/api/app/users/Users.go` 中 `UpdateAuth`，调用 `user.UserCubeStillInOrganizersTeam`。

---

## 5. 主办团队接口

### 5.1 列表（已有）

- **GET** `/v3/cube-api/admin/competition/organizers`  
- 行为：分页列表，与现有 `AllOrganizers` 一致。

### 5.2 创建主办团队

- **POST** `/v3/cube-api/admin/competition/organizers`

**请求体**：

```json
{
  "name": "团队名称",
  "introduction": "",
  "email": "",
  "leader_cube_id": "",
  "status": "Using",
  "leader_remark": "",
  "admin_message": "",
  "ass_cube_ids": ["CUBE01", "CUBE02"]
}
```

| 字段 | 必填 | 说明 |
|------|------|------|
| `name` | 是 | 去空格后非空；全局唯一。 |
| `status` | 是 | 须为合法 `OrganizersStatus`。 |
| `leader_cube_id` | 否 | 若填写则对应用户须存在。 |
| `ass_cube_ids` | 否 | 成员 CubeID；会去重、校验用户存在；与组长重复项会忽略。 |

**响应**：创建后的 `Organizers` 对象。

### 5.3 主办团队详情（含群组列表）

- **GET** `/v3/cube-api/admin/competition/organizers/:orgId`

**响应示例**：

```json
{
  "organizer": { },
  "groups": [ ]
}
```

### 5.4 更新主办团队

- **PUT** `/v3/cube-api/admin/competition/organizers/:orgId`

**请求体**（均为可选，仅传入需要修改的字段）：

```json
{
  "name": "新名称",
  "introduction": "",
  "email": "",
  "leader_cube_id": "",
  "status": "Using",
  "leader_remark": "",
  "admin_message": "",
  "ass_cube_ids": ["CUBE01"]
}
```

说明：传 `ass_cube_ids` 时表示 **整体替换** 成员列表（不含组长；组长单独用 `leader_cube_id`）。

### 5.5 删除主办团队

- **DELETE** `/v3/cube-api/admin/competition/organizers/:orgId`

**规则**：若仍存在 `orgId` 指向该团队的 **比赛（Competition）** 记录，则拒绝删除。  
删除成功时，会删除该团队下全部 **比赛群组** 记录（硬删除群组表行）。

---

## 6. 比赛群组接口

### 6.1 某主办下的群组列表

- **GET** `/v3/cube-api/admin/competition/organizers/:orgId/groups`  
- 支持通用列表查询参数（与 `GenerallyList` 一致：`page`、`size` 等）。

### 6.2 创建比赛群组

- **POST** `/v3/cube-api/admin/competition/organizers/:orgId/groups`

**请求体**：

```json
{
  "name": "群组名",
  "qq_groups": ["123456", "789012"],
  "qq_group_uid": ["uid1"],
  "wechat_groups": []
}
```

**规则**：同一 `orgId` 下已有群组数 **≥ 3** 时返回校验错误。

### 6.3 更新比赛群组

- **PUT** `/v3/cube-api/admin/competition/groups/:groupId`

**请求体**（可选字段）：

```json
{
  "name": "新名称",
  "qq_groups": ["123456"],
  "qq_group_uid": [],
  "wechat_groups": []
}
```

### 6.4 删除比赛群组

- **DELETE** `/v3/cube-api/admin/competition/groups/:groupId`

---

## 7. 主办成员（主办管理员）

### 7.1 新增成员

- **POST** `/v3/cube-api/admin/competition/organizers/:orgId/members`

**请求体**：

```json
{
  "cube_id": "目标用户 CubeID",
  "grant_auth": true
}
```

| 字段 | 说明 |
|------|------|
| `grant_auth` | 为 `true` 时，在保存成员后为用户 **置位** `AuthOrganizers`（主办权限）。 |

**规则**：不能将组长重复添加为成员；用户必须存在。

### 7.2 移出成员

- **DELETE** `/v3/cube-api/admin/competition/organizers/:orgId/members?cube_id=xxx`

**规则**：不能移出 **组长**；需先将 `leader_cube_id` 改为他人或清空。

---

## 8. 兼容接口

- **POST** `/v3/cube-api/admin/competition/:orgId`  
  仍为 `DoWithOrganizers`：按请求体更新状态、管理员留言等（与历史行为一致）。

---

## 9. 错误码与 HTTP 状态

常用封装见 `src/api/exception/code.go`：

| 场景 | 典型错误 |
|------|-----------|
| 参数不合法 | `10008` 无效输入（`ErrInvalidInput`） |
| 校验失败（含主办权限撤销冲突、群组超限） | `10014`（`ErrValidationFailed`） |
| 资源不存在 | `10013`（`ErrResourceNotFound`） |
| 无权限（非超级管理员） | `10004`（`ErrAuthField`） |

---

## 10. 源码索引（维护用）

| 内容 | 路径 |
|------|------|
| 管理端路由 | `src/api/routes/admin.go` |
| 主办/群组 Handler | `src/api/app/organizers/admin_manage.go` |
| 用户权限更新 | `src/api/app/users/Users.go`（`UpdateAuth`） |
| 主办模型与常量 | `src/internel/database/model/user/sponsor.go` |
| 比赛群组模型 | `src/internel/database/model/competition/compertion_group.go` |
| 多值字段序列化 | `src/internel/database/model/competition/group_fields.go` |
| 成员是否仍绑定团队 | `user.UserCubeStillInOrganizersTeam`（`sponsor.go`） |

---

## 11. 前端对接建议

1. 所有写操作在 UI 上二次确认（尤其删除团队、删除群组、撤销主办权限）。  
2. 撤销主办权限前，先调成员删除或更新团队接口，确保用户不在 `leader` / `ass_cube_ids` 中。  
3. 多值群字段统一用数组提交；展示时直接绑定接口返回的数组（若后端在部分场景返回字符串，可按文档约定兼容）。  
4. 列表接口分页字段与项目其他 `GenerallyList` 页面保持一致。
