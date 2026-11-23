# 群赛后端 API 契约

所有路径均相对于 `/v3/cube-api`。需要登录的请求使用 `Authorization: Bearer <token>`。
成功响应为 `{"code":"200","msg":"ok","data":...}`；失败沿用现有错误封装，详情在 `error`，不要只判断 HTTP 200 或假设错误响应含 `msg`。

## 业务边界与 HEAD 恢复

群赛唯一边界为 `Genre == OnlineInformal && GroupID > 0`（OnlineInformal=5）。现有创建入口在 HEAD 中固定保存 OnlineInformal，GroupID 是它与无群线上比赛的区别；不根据请求 genre 改比赛类型。
以下群赛资格、数量校验、批量解析、待审覆盖及轮次管理规则仅作用于该边界。最新规则：群赛成绩操作不受日期、IsDone、ResultClosedAt 或轮次开放/关闭时间限制；时间函数只用于展示，不作为成绩写入许可判断。
WCA/Official/OnlineOfficial/Informal，以及 GroupID=0 的 OnlineInformal 均走 HEAD 业务分支，哪怕其他 Genre 存有 GroupID 也不算群赛。

| 流程 | 非群赛恢复路径 |
| --- | --- |
| IsRunningTime | 保留 HEAD 的当前本地日期前后日判断；不引入北京比赛整日或 IsDone 校验，原有时间计算缺陷本轮也不修复 |
| 直接录入 | LegacyAddResult：轮次查询不强制 IsRunning；OnlineInformal 原有自动报名/追加项目；其他类型保留报名、晋级和上一轮成绩要求；保留原截断/补齐成绩方式 |
| Web 预成绩 | LegacySavePending：先检查原有报名和项目资格，不自动报名；不强制开放轮次，不加上一轮成绩检查；按原 round 文本查询覆盖。仅保留实际模型保存及 CubeID/RoundNumber 必要数据修复 |
| 预成绩审批 | legacyApprove：使用原当前时间及成绩资格路径，timeout 不改成终态，不套历史截止/时间证据规则；仍强制 URL 比赛对象归属 |
| refresh_event | legacyRefresh：保留原开放/关闭判断、晋级严格小于人数边界，以及已有下一轮状态；不增加互斥/待审阻止/实际时间戳规则 |
| end | 仅按 HEAD 设置 IsDone，不读取结束时钟，不关闭轮次或写 ResultClosedAt |
| 编辑 | 不验证新增日期规则，不接管轮次状态，不增加 CanPreResult 编辑行为 |
| 创建无群比赛 | 仍固定 OnlineInformal，默认 Reviewing、CanPreResult=true；保留传入轮次状态，不强制手动关闭 |
| organizer 比赛列表/非群赛预成绩列表 | 恢复 HEAD 的原列表查询和返回形式；管理员群赛列表仍使用独立白名单查询 |
| 机器人非群赛类型 | 选中非群赛类型后使用原逐行解析和原模型保存，不套新批量、多盲校验、覆盖和资格规则 |

必要对象归属、JWT、所有者删除、保存错误和事务保护继续保留；不涉及公式库、配置或用户权限外围重构。

## 管理员群赛管理

管理员和超级管理员任一权限即可使用下表接口，无需玩家位、主办位或主办团队成员身份。后台按 URL 中的比赛加载实际主办团队；`/comps/:compId` 仅管理上述群赛边界。
普通 organizer 接口的团队权限不变。
既有管理员 DELETE 比赛接口保留 HEAD 对非群赛的删除能力，新增管理读写接口仅支持群赛。

| Method | Path | JSON body / query | data |
| --- | --- | --- | --- |
| GET | `/admin/competition/approvals/comps` | 下方列表筛选 | `{items,total,next}` |
| GET | `/admin/competition/filters` | page/size，见下方选项契约 | 团队、群组选项及分页元数据 |
| POST | `/admin/competition/approvals/:compId/approval` | `{"Ok":true}` 或 `{"Ok":false,"RejectMsg":"拒绝原因"}` | null |
| GET | `/admin/competition/comps/:compId` | 无 | Competition 详情，补 OrganizersName/GroupName，包含 comp_json 和管理所需打乱 |
| POST | `/admin/competition/comps/:compId` | 沿用 organizer 完整编辑 body；项目字段为 `comp_json` | 更新后的 Competition |
| POST | `/admin/competition/comps/:compId/update_name` | `{"newName":"名称"}` | null |
| POST | `/admin/competition/comps/:compId/apply` | 无 | null |
| POST | `/admin/competition/comps/:compId/end` | 无 | null |
| DELETE | `/admin/competition/comps/:compId` | 无 | null |
| GET | `/admin/competition/comps/:compId/players` | 无 | Registration 数组 |
| POST | `/admin/competition/comps/:compId/players/approval/:reg_id` | `{"pass":true}` 或 `{"pass":false}` | null |
| GET | `/admin/competition/comps/:compId/result` | 原参数 `event_id,round_num` 或 `cube_id` | Results 数组 |
| POST | `/admin/competition/comps/:compId/result` | `{"Results":[1,2,3,4,5],"CubeID":"选手CubeID","Round":1,"EventId":"333","Penalty":[]}` | null |
| DELETE | `/admin/competition/comps/:compId/result/:result_id` | 无 | null |
| GET | `/admin/competition/comps/:compId/pre_results` | 下方预成绩筛选 | `{items,total,next}` |
| POST | `/admin/competition/comps/:compId/pre_results/:result_id/approval` | `{"FinishDetail":"ok"}` | null |
| PUT | `/admin/competition/comps/:compId/refresh_event` | 单轮 `{"EventID":"333","RoundNumber":1,"Open":true}`；全部开放 `{"All":true,"Open":true}` | null |

`Ok`、`pass`、`Open` 必须明确传布尔值，false 有效。比赛审核仅接受 Reviewing 状态；拒绝必须有非空 RejectMsg，通过时清空旧拒绝原因。
预成绩 FinishDetail 支持 `ok/not/wait/doubtful/timeout`；ok/not/timeout 为终态，wait/doubtful 保持待审。终态不能重复处理。
成绩删除、预成绩审批和报名审批都绑定 URL 比赛，不能用其他比赛的记录 ID。

### 旧人工确认参数兼容

`POST /admin/competition/comps/:compId/pre_results/:result_id/approval` 仍兼容旧可选 JSON 字段，但不再要求人工确认或时间证据；普通 `{"FinishDetail":"ok"}` 即可审批赛前、赛后、已结束或闭轮群赛成绩：

```json
{
  "FinishDetail": "ok",
  "HistoricalConfirmed": true,
  "ReviewReason": "可选备注"
}
```

- 仅管理员/超级管理员专用路由允许确认，服务层同时检查管理员角色；即使登录账号是管理员，通过 organizer 路由提交 HistoricalConfirmed=true 也明确拒绝。
- HistoricalConfirmed 默认 false，传 true 也不改变业务校验；不检查时间记录是否完整，ReviewReason 不再必填。成功 data=null，沿用现有成功封装。
- 未审批、跨比赛、身份/项目/轮次归属、晋级资格及成绩格式错误仍拒绝；赛期、截止、轮次时间缺失或超出均不再拒绝。
- 不改写 CreatedAt/UpdatedAt、轮次时间或 ResultClosedAt，不为旧赛自动编造截止。
- 兼容参数传 true 时，Processor/ProcessorID 记录实际操作者，Detail 保存 `状态; HistoricalConfirmed=true; ReviewReason="备注"`，有旧 Detail 时附带 PreviousDetail；Finish/FinishDetail/ResultID 仍按实际审批状态保存。前端审批结果筛选继续依赖 FinishDetail，不能按 Detail==ok 判断。

## 比赛列表

`GET /admin/competition/approvals/comps` query 全部为白名单字段，组合条件按 AND 执行：

| query | 类型 | 含义 |
| --- | --- | --- |
| id | 正整数 | 比赛 ID 精确匹配 |
| name | string | 名称包含查询 |
| status | string | Reviewing/Running/Reject/Temporary/Ban；省略默认 Reviewing；`all` 不限状态 |
| orgId | 正整数 | 所属团队 ID |
| groupId | 正整数 | 所属群组 ID |
| isDone | true/false | 是否显式结束；省略不限 |
| date | YYYY-MM-DD | 比赛日期覆盖该北京时间日期 |
| page | int | 从 1 起，默认 1 |
| size | int | 默认 20，最大 100 |

data 为 `{items:[],total:0,next:false}`。按 `created_at DESC,id DESC` 排序。
每项字段固定为：`id,createdAt,updatedAt,StrId,Name,Status,RejectMsg,IsDone,OrganizersID,OrganizersName,GroupId,GroupName,CompStartTime,CompEndTime,EventMin,Genre,CanPreResult,Count,PendingCount`。
PendingCount 统计该比赛所有未完成审批预成绩。列表不返回 comp_json、Scrambles 或打乱文件路径。
`GET /organizers/:orgId/comp/` 为兼容其他线上类型，恢复 HEAD 的所属团队原列表，不使用本节新增筛选及分页契约。
不支持通用 `search/like/order/include_deleted` 绕过此范围。

### 管理员只读筛选选项和详情名称

`GET /admin/competition/filters?page=1&size=200` 位于普通 admin 路由，管理员和超级管理员都能访问，独立于 super-only 团队编辑接口。返回现有未软删除的团队和群组真实名称，无比赛状态过滤：

```json
{
  "code": "200",
  "msg": "ok",
  "data": {
    "organizers": [{"id": 18, "Name": "团队名称"}],
    "groups": [{"id": 9, "Name": "群组名称", "OrganizersID": 18}],
    "page": 1,
    "size": 200,
    "organizersTotal": 1,
    "groupsTotal": 1,
    "organizersNext": false,
    "groupsNext": false
  }
}
```

page 默认 1，size 默认 200、上限 500；两组独立按 id ASC 使用同一 page/size 分页。前端累积选项，直到两个 next 均为 false，避免静默截断；可以按 groups.OrganizersID 在前端联动团队选项。
管理员详情在 Competition 原结构上补 OrganizersName 和 GroupName，不修改数据库模型；列表与详情共用名称解析。软删除/缺失关联显示“主办团队不存在（ID 18）”或“群组不存在（ID 9）”，空名称显示“未命名…”提示。缺失团队的管理员 GET 详情可读，写操作仍要求团队存在。

## 预成绩筛选

管理员和 organizer 的群赛 `GET .../:compId/pre_results` 都支持下表；非群赛 organizer 预成绩列表保持原行为：

| query | 类型 | 含义 |
| --- | --- | --- |
| finish | false/true/all | 默认 false，待审/已完成/全部 |
| finishDetail | ok/not/timeout | 精确筛选终态 FinishDetail；省略 finish 时自动按 true 查询 |
| review_status | passed/rejected | passed 等价 finishDetail=ok；rejected 等价 finishDetail=not，不包含 timeout |
| event | string | EventID 精确匹配 |
| round | int | RoundNumber 精确匹配 |
| player | string | PersonName 包含或 CubeID 精确匹配 |
| source | string | Source 精确匹配，如 web-api 或 robot-group-7 |
| start_time | Unix 秒 | created_at 包含该起点 |
| end_time | Unix 秒 | created_at 不含该终点 |
| page/size | int | 默认 1/20，size 最大 100 |

按 `created_at ASC,id ASC` 返回 `{items,total,next}`。
`total` 是全部筛选条件生效后的总数（分页之前），`next` 由该总数计算。终态筛选自动附加 finish=true；与 finish=false 或相互冲突的 finishDetail/review_status 同传时返回参数错误。可直接请求 `?review_status=passed&page=1&size=20` 或 `?review_status=rejected&page=1&size=20`，无需逐页获取全部记录后前端过滤。非法枚举不进入查询。
items 沿用 PreResults 字段，主要为 `id,createdAt,updatedAt,CompetitionID,CompetitionName,Round,RoundNumber,PersonName,UserID,CubeID,Result,Penalty,EventID,EventName,EventRoute,Best,Average,ResultID,CompsName,RoundName,Recorder,Processor,ProcessorID,Finish,Detail,FinishDetail,Source`。
Results 的部分字段沿用 omitempty，零值可能省略；PreResults 自有字段沿用首字母大写。

## 玩家 Web 预录入

`/player_comp` 全部先执行 JWT，再检查玩家权限。

| Method | Path | body | data |
| --- | --- | --- | --- |
| POST | `/player_comp/result/pre/add_results` | `{"CompId":1,"Round":"1","Results":[1,2,3,4,5],"Penalty":[],"EventID":"333"}` | 已保存的 PreResults（含 id） |
| DELETE | `/player_comp/result/pre/delete/:pre_id` | 无 | null |
| GET | `/player_comp/result/pre/list` | 原列表 query | 原个人列表封装 |

同一比赛/选手/项目/轮次的未完成预成绩覆盖原记录；已完成记录保留，再提交会新增待审记录。
只允许删除自己的未完成记录；所有者和 finish 条件在事务内再次验证。
仅群赛首个合法提交可自动注册选手/追加项目，已有被拒绝或未通过报名不会自动转为通过。其他线上类型 Web 预成绩仍要求事先报名。

## 日期、审批与轮次

- 群赛新增、待审覆盖、直接成绩新增/修改、审批及删除仅要求比赛已审批 Status=Running，不受赛前/赛后、IsDone、ResultClosedAt、IsRunning、ActualStartTime/ActualEndTime、后续轮次已开放时间限制。管理页已关闭轮次也可补录。
- 仍保留权限、比赛/群/用户/项目/轮次归属、CanPreResult（预录入及预成绩通过）、成绩格式、晋级名单及上一轮成绩资格、事务和并发保护。已有后续轮次实际成绩时仍不可删除前轮成绩，避免破坏原上一轮资格依赖。
- end 允许存在待审成绩，记录 comp_json.ResultClosedAt 并设置 IsDone=true、关闭轮次；这些是状态展示与机器人默认选择元数据，不关闭成绩操作能力。
- IsRunningTimeAt 的北京时间整日判断保留作展示，成绩服务不依赖该函数；不再要求任何历史时间证据确认。
- 手动关闭单个轮次仍要求处理该轮待审成绩；显式 end 不受该限制。
- 预成绩缺少 UserID/CubeID/RoundNumber、选手不匹配、项目轮次不存在或资格不合法时，ok 明确报错，不静默补猜；提交时间不参与许可判断。
- 审批使用预成绩提交/最后修改时间，不把当前审批时间当新提交时间；审批状态修改保留 updated_at 作为最后内容修改时间。
- 线上群赛创建时全部轮次默认开放；审批通过不改轮次状态，已有轮次仍可通过 refresh_event 调整。
- 支持一键开放全部轮次：管理员和主办复用现有 PUT refresh_event 路由，发送 `{"All":true,"Open":true}`。只支持 Genre=5 且 GroupID>0 的已批准 Running、IsDone=false 群赛，不依赖日期，赛前也可操作。无需新增路由或权限。
- All=true 必须明确 Open=true，不允许同时提供 EventID/RoundNumber（包括空字符串或0）；不支持一键关闭。成功返回 `{"code":"200","msg":"ok","data":null}`，冲突请求返回参数错误，非群赛拒绝。
- 一键操作在一次事务及比赛行锁内设置全部轮 IsRunning=true，保留已有 ActualStartTime，缺失则设本次时间，清空 ActualEndTime；不改资格、晋级名单、打乱或成绩。重复调用不刷新原开放时间，已全部开放且时间完整时不写数据库。
- 群赛单轮开放同步取消日期、同项目互斥和前后轮开放顺序限制，允许同一项目多轮同时开放。仍拒绝未审批或 IsDone=true；关闭轮次的待审检查、晋级计算及原非群赛逻辑保持不变。这些赛程管理规则不影响成绩补录或修改权限。
- 关闭轮次计算下一轮晋级名单，名次等于人数边界也晋级，同名次保留；Competitors=0 不限制人数。
- organizer 正确轮次路径为 `PUT /organizers/:orgId/comp/:compId/refresh_event`。原重复 compId 的异常路径已移除。
- 编辑比赛 JSON 不允许通过 IsRunning/ActualStartTime/ActualEndTime/AdvancedToNextRound 绕过轮次管理，保留已有服务端状态，新轮次默认开放。
- ResultClosedAt 仅由服务端 end 写入；创建时清空，编辑时保留原值，重复 end 不延后截止。该字段保存在现有 comp_json 列，无需数据库 schema 迁移。
- end 接收时钟函数，在取得比赛行锁后才读取当前时间，仅用于准确记录结束元数据。

## 创建与申请

`POST /organizers/:orgId/comp/` 保持原 body：`Name,StrId,Illustrate,IllustrateHTML,Location,Country,City,RuleMD,RuleHTML,CompJSON,Count,CanPreResult,CompStartTime,CompEndTime,GroupID,CanStartedAddEvent,UseTNoodle,Apply`。
GroupID 必须是该团队拥有的有效群组；日期必须有效。
Apply=false 保存 Temporary；Apply=true 保存 Reviewing。CanPreResult 按请求值保存。
`POST .../:compId/apply` 将 Temporary/Reject 转 Reviewing，并清空旧拒绝原因。
以上为 GroupID>0 群赛规则。GroupID=0 保留 HEAD 创建行为；传入 genre 不会改变现有入口固定 OnlineInformal 的逻辑。

## 机器人

- 复用 StringListFromDB 精确匹配群 ID，兼容 JSON 数组和历史逗号列表；CQHTTP QQ 使用 qq_groups，官方 QQBot 使用 qq_group_uid，不混用两个平台命名空间。当前 InMessage 无微信平台身份字段。
- 群赛默认候选只取本群 Status=Running、IsDone=false、CanPreResult=true 的群赛，忽略日期，按 created_at ASC,id ASC 选择首个。选定后项目/轮次/资格不合法则报错，不跳到另一场。
- `录入-ID ...` 只查本群指定比赛；已审批且允许预录入的群赛即使已结束、过期或尚未到赛期也可录入。不存在、未审批、禁预录入或跨群时失败，不回退。
- 本群没有未结束的默认群赛时，提示使用“录入-ID”指定历史比赛，不自动转投最早历史群赛。
- 批量使用换行或两侧有空白的 `/`；多盲使用 `3/4 10:00`，分数内 `/` 不拆批。多次多盲可连续提供多个三元组。
- 批量任一项不合法或保存失败，全部成绩和自动报名回滚；保存成功后才回复录入成功。

## 验证范围与限制

使用隔离 MySQL 数据库创建测试表，不读取业务配置或生产 DSN，不执行旧集成测试。设置 `CUBING_TEST_MYSQL_DSN` 指向测试服务器，夹具自动创建并清理临时数据库；未配置时跳过数据库测试。实现使用 GORM 事务及比赛行 FOR UPDATE。
不新增依赖，不修改生产 schema，不修复历史坏记录。群组历史重复映射会拒绝机器人录入，需人工清理。

已通过的验证命令：

```sh
go test -race ./src/internel/groupcomp ./src/internel/database/model/competition
go build ./src/api/routes ./src/robot/robots/plugin
go vet ./src/internel/groupcomp ./src/internel/database/model/competition
```

API 路由检查使用 AST/源码契约测试，不启动 HTTP 服务或导入有后台任务 init 的 svc；尚未执行真实 JWT HTTP 端到端联调和 MySQL 并发压测。
服务流程测试覆盖赛前、赛后、显式结束、轮次关闭、后续轮次已开、时间字段缺失或冲突情况下新增/覆盖/修改/审批/删除成功；未审批、跨比赛、格式/资格不合法仍拒绝，机器人默认和指定 ID 分开选择。
旧人工确认字段可选兼容且不再强制原因；路由权限和操作者记录仍保留。结束时钟测试仅验证元数据读取顺序，不替代真实 MySQL 行锁调度压测。
群赛与其他线上类型回归覆盖非群赛 HEAD 时间语义、闭轮录入/截断、Web 报名与覆盖、审批 timeout、轮次/结束状态、群赛仍执行新校验、管理员筛选选项分页与名称一致性。
一键全开测试覆盖全关/部分开放、赛前开放、一次锁、时间保留及幂等、未审批/已结束/非群赛拒绝、保存失败事务无部分变更、资格/打乱/预成绩不变，以及单轮多轮同时开放。
