# 王双媛举报申诉处理中状态开发说明

## 一、改动目标

本次补充举报和申诉的受理动作，让管理员可以先把待处理记录标记为处理中，再进行最终处理。

状态流转如下：

```text
PENDING -> PROCESSING -> RESOLVED / REJECTED / CLOSED
PENDING -> PROCESSING -> APPROVED / REJECTED / CLOSED
```

说明：

- 举报最终状态使用 `RESOLVED`、`REJECTED`、`CLOSED`。
- 申诉最终状态使用 `APPROVED`、`REJECTED`、`CLOSED`。
- `PROCESSING` 表示管理员已经接手处理，普通用户不能修改、撤回或重复提交同一目标的举报/申诉。

## 二、新增接口

### 2.1 管理员标记举报为处理中

```http
POST /admin/reports/:id/processing
```

权限：

```text
登录 + role=ADMIN
```

业务规则：

- 只有 `PENDING` 状态的举报可以标记为 `PROCESSING`。
- 标记成功后写入 `handler_id`。
- 标记成功后写入管理员操作日志。
- 如果举报已经是 `PROCESSING`、`RESOLVED`、`REJECTED` 或 `CLOSED`，返回业务错误。

### 2.2 管理员标记申诉为处理中

```http
POST /admin/appeals/:id/processing
```

权限：

```text
登录 + role=ADMIN
```

业务规则：

- 只有 `PENDING` 状态的申诉可以标记为 `PROCESSING`。
- 标记成功后写入 `handler_id`。
- 标记成功后写入管理员操作日志。
- 如果申诉已经是 `PROCESSING`、`APPROVED`、`REJECTED` 或 `CLOSED`，返回业务错误。

## 三、用户侧限制

用户侧没有新增修改接口。

用户仍然只能：

- 提交举报或申诉。
- 查看自己的举报或申诉列表。
- 查看自己的举报或申诉详情。

当记录进入 `PROCESSING` 后：

- 用户不能重复提交同一目标的举报。
- 用户不能重复提交同一目标的申诉。
- 用户不能处理、修改、删除或撤回记录。

后端已有重复限制：

```text
reports  : 同一 reporter_id + target_type + target_id 在 PENDING / PROCESSING 下不能重复提交
appeals  : 同一 appellant_id + target_type + target_id 在 PENDING / PROCESSING 下不能重复提交
```

## 四、涉及代码

```text
backend/internal/report/repository.go
backend/internal/report/service.go
backend/internal/report/handler.go
backend/internal/report/router.go

backend/internal/appeal/repository.go
backend/internal/appeal/service.go
backend/internal/appeal/handler.go
backend/internal/appeal/router.go
```

## 五、前端建议

列表页根据状态展示操作按钮：

```text
PENDING      管理员显示“开始处理”
PROCESSING   管理员显示“处理结果”，用户只显示状态
最终状态      只显示结果，不显示处理按钮
```

普通用户看到 `PROCESSING` 时，只展示“处理中”，不要展示撤回、修改、再次提交等按钮。
