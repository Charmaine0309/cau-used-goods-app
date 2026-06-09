# 王双媛 举报申诉状态流转与管理员后续处理接口测试说明

## 一、测试范围

本测试说明覆盖以下接口：

```http
POST /reports/:id/close
POST /appeals/:id/close
POST /admin/reports/:id/processing
POST /admin/reports/:id/handle
POST /admin/appeals/:id/processing
POST /admin/appeals/:id/handle
PUT /admin/users/:id/status
PUT /admin/products/:id/status
PUT /admin/orders/:id/status
```

重点验证：

- 用户只能在 `PENDING` 关闭举报或申诉。
- 管理员不能处理为 `CLOSED`。
- 管理员只能处理为 `APPROVED` 或 `REJECTED`。
- 处理完成后发送站内消息。
- 后续商品、订单、账号状态由管理员接口单独修改。
- 管理员状态修改接口会写入 `admin_logs`。

## 二、测试前置条件

服务地址：

```powershell
$baseUrl = "http://localhost:8080"
```

建议配置：

```yaml
server:
  env: dev
  port: 8080
```

准备 Token：

```powershell
$userToken = "普通用户 JWT"
$adminToken = "管理员 JWT"
```

请求头：

```powershell
$userHeaders = @{
  Authorization = "Bearer $userToken"
  "Content-Type" = "application/json"
}

$adminHeaders = @{
  Authorization = "Bearer $adminToken"
  "Content-Type" = "application/json"
}
```

说明：

- 普通用户需要已登录。
- 举报接口需要学生认证。
- 管理员接口需要 `role = ADMIN`。
- 以下 `$reportId`、`$appealId`、`$userId`、`$productId`、`$orderId` 需要替换为实际测试数据 ID。

## 三、举报关闭测试

### 3.1 用户关闭 PENDING 举报

```powershell
$reportId = 1

Invoke-RestMethod `
  -Method POST `
  -Uri "$baseUrl/reports/$reportId/close" `
  -Headers $userHeaders `
  -Body '{"closeReason":"用户主动撤回举报"}'
```

预期：

```text
HTTP 200
status = CLOSED
handlerId 为空
```

数据库检查：

```sql
SELECT id, status, handle_result, handler_id, handle_time
FROM reports
WHERE id = 1;
```

### 3.2 用户关闭 PROCESSING 举报应失败

先由管理员接手：

```powershell
Invoke-RestMethod `
  -Method POST `
  -Uri "$baseUrl/admin/reports/$reportId/processing" `
  -Headers $adminHeaders
```

用户再关闭：

```powershell
Invoke-WebRequest `
  -Method POST `
  -Uri "$baseUrl/reports/$reportId/close" `
  -Headers $userHeaders `
  -Body '{"closeReason":"处理中尝试撤回"}'
```

预期：

```text
HTTP 400
message = report cannot be closed
status 仍为 PROCESSING
```

## 四、申诉关闭测试

### 4.1 用户关闭 PENDING 申诉

```powershell
$appealId = 1

Invoke-RestMethod `
  -Method POST `
  -Uri "$baseUrl/appeals/$appealId/close" `
  -Headers $userHeaders `
  -Body '{"closeReason":"用户主动撤回申诉"}'
```

预期：

```text
HTTP 200
status = CLOSED
handlerId 为空
```

数据库检查：

```sql
SELECT id, status, handle_result, handler_id, handle_time
FROM appeals
WHERE id = 1;
```

### 4.2 用户关闭 PROCESSING 申诉应失败

先由管理员接手：

```powershell
Invoke-RestMethod `
  -Method POST `
  -Uri "$baseUrl/admin/appeals/$appealId/processing" `
  -Headers $adminHeaders
```

用户再关闭：

```powershell
Invoke-WebRequest `
  -Method POST `
  -Uri "$baseUrl/appeals/$appealId/close" `
  -Headers $userHeaders `
  -Body '{"closeReason":"处理中尝试撤回"}'
```

预期：

```text
HTTP 400
message = appeal cannot be closed
status 仍为 PROCESSING
```

## 五、管理员处理举报测试

### 5.1 管理员处理为 APPROVED

```powershell
Invoke-RestMethod `
  -Method POST `
  -Uri "$baseUrl/admin/reports/$reportId/handle" `
  -Headers $adminHeaders `
  -Body '{"status":"APPROVED","handleResult":"举报成立，后续将下架商品"}'
```

预期：

```text
HTTP 200
status = APPROVED
handlerId = 当前管理员 ID
handleTime 非空
```

数据库检查：

```sql
SELECT id, status, handle_result, handler_id, handle_time
FROM reports
WHERE id = 1;

SELECT id, receiver_id, sender_id, message_type, related_type, related_id
FROM messages
WHERE related_type = 'REPORT' AND related_id = 1
ORDER BY id DESC;

SELECT id, operation_type, target_type, target_id, description
FROM admin_logs
WHERE target_type = 'REPORT' AND target_id = 1
ORDER BY id DESC;
```

### 5.2 管理员处理为 CLOSED 应失败

```powershell
Invoke-WebRequest `
  -Method POST `
  -Uri "$baseUrl/admin/reports/$reportId/handle" `
  -Headers $adminHeaders `
  -Body '{"status":"CLOSED","handleResult":"管理员尝试关闭"}'
```

预期：

```text
HTTP 400
请求体校验失败或 message = status must be APPROVED or REJECTED
```

### 5.3 管理员处理为 RESOLVED 应失败

```powershell
Invoke-WebRequest `
  -Method POST `
  -Uri "$baseUrl/admin/reports/$reportId/handle" `
  -Headers $adminHeaders `
  -Body '{"status":"RESOLVED","handleResult":"旧状态测试"}'
```

预期：

```text
HTTP 400
不允许 RESOLVED
```

## 六、管理员处理申诉测试

### 6.1 管理员处理为 APPROVED

```powershell
Invoke-RestMethod `
  -Method POST `
  -Uri "$baseUrl/admin/appeals/$appealId/handle" `
  -Headers $adminHeaders `
  -Body '{"status":"APPROVED","handleResult":"申诉通过，后续将恢复商品"}'
```

预期：

```text
HTTP 200
status = APPROVED
handlerId = 当前管理员 ID
handleTime 非空
```

数据库检查：

```sql
SELECT id, status, handle_result, handler_id, handle_time
FROM appeals
WHERE id = 1;

SELECT id, receiver_id, sender_id, message_type, related_type, related_id
FROM messages
WHERE related_type = 'APPEAL' AND related_id = 1
ORDER BY id DESC;

SELECT id, operation_type, target_type, target_id, description
FROM admin_logs
WHERE target_type = 'APPEAL' AND target_id = 1
ORDER BY id DESC;
```

### 6.2 申诉 APPROVED 后不自动恢复商品或账号

如果申诉目标是商品：

```sql
SELECT target_type, target_id
FROM appeals
WHERE id = 1;

SELECT id, status, is_deleted
FROM products
WHERE id = <target_id>;
```

预期：

```text
商品状态不会因为申诉 APPROVED 自动变化
```

如果申诉目标是用户：

```sql
SELECT id, account_status, is_deleted
FROM users
WHERE id = <target_id>;
```

预期：

```text
账号状态不会因为申诉 APPROVED 自动变化
```

### 6.3 管理员处理为 CLOSED 应失败

```powershell
Invoke-WebRequest `
  -Method POST `
  -Uri "$baseUrl/admin/appeals/$appealId/handle" `
  -Headers $adminHeaders `
  -Body '{"status":"CLOSED","handleResult":"管理员尝试关闭"}'
```

预期：

```text
HTTP 400
message = status must be APPROVED or REJECTED
```

## 七、管理员修改账号状态测试

### 7.1 禁用账号

```powershell
$userId = 2
$userReportId = 1 # 替换为 target_type=USER,target_id=$userId 的举报 ID

Invoke-RestMethod `
  -Method PUT `
  -Uri "$baseUrl/admin/users/$userId/status" `
  -Headers $adminHeaders `
  -Body "{`"accountStatus`":`"DISABLED`",`"reason`":`"举报成立，禁用账号`",`"relatedType`":`"REPORT`",`"relatedId`":$userReportId}"
```

预期：

```text
HTTP 200
accountStatus = DISABLED
```

数据库检查：

```sql
SELECT id, account_status, is_deleted
FROM users
WHERE id = 2;

SELECT id, operation_type, target_type, target_id, description, related_type, related_id
FROM admin_logs
WHERE operation_type = 'USER_DISABLE'
ORDER BY id DESC;
```

### 7.2 存在待确认订单时禁用账号应自动异常关闭订单

准备条件：目标用户作为卖家或买家存在 `PENDING_CONFIRM` 订单，且不存在 `WAIT_MEET` 订单。

```powershell
Invoke-WebRequest `
  -Method PUT `
  -Uri "$baseUrl/admin/users/$userId/status" `
  -Headers $adminHeaders `
  -Body '{"accountStatus":"DISABLED","reason":"待确认订单自动异常关闭测试","relatedType":"REPORT","relatedId":1}'
```

预期：

```text
HTTP 200
用户 account_status = DISABLED
相关 PENDING_CONFIRM 订单变为 EXCEPTION_CLOSED
订单 cancel_reason 已记录，close_time 非空
相关订单 admin_logs 写入 UPDATE_ORDER_STATUS/ORDER_EXCEPTION_CLOSE，并继承 related_type=REPORT、related_id=1
相关商品按角色处理：
- 被禁用用户是买家：商品恢复为 ON_SALE
- 被禁用用户是卖家：相关锁定商品下架
被禁用用户原有 ON_SALE 商品下架
系统消息在事务提交后生成，订单消息 related_type=ORDER，账号消息 related_type=USER
```

注意：`relatedId=1` 只是示例，实际测试必须替换为目标用户匹配的举报 ID；否则关联来源校验会失败。

### 7.3 存在待面交订单时禁用账号应失败

准备条件：目标用户作为卖家或买家存在 `WAIT_MEET` 订单。

```powershell
Invoke-WebRequest `
  -Method PUT `
  -Uri "$baseUrl/admin/users/$userId/status" `
  -Headers $adminHeaders `
  -Body '{"accountStatus":"DISABLED","reason":"待面交订单阻断测试"}'
```

预期：

```text
HTTP 400
message 包含待面交订单明细，例如：
该用户作为卖家存在1个待面交订单，请先处理待面交订单后再修改账号状态
用户状态不变
订单状态不变
商品状态不变
```

### 7.4 恢复账号

```powershell
$userAppealId = 1 # 替换为 target_type=USER,target_id=$userId 的申诉 ID

Invoke-RestMethod `
  -Method PUT `
  -Uri "$baseUrl/admin/users/$userId/status" `
  -Headers $adminHeaders `
  -Body "{`"accountStatus`":`"NORMAL`",`"reason`":`"申诉通过，恢复账号`",`"relatedType`":`"APPEAL`",`"relatedId`":$userAppealId}"
```

预期：

```text
HTTP 200
accountStatus = NORMAL
is_deleted = 0
```

### 7.5 非法账号状态应失败

```powershell
Invoke-WebRequest `
  -Method PUT `
  -Uri "$baseUrl/admin/users/$userId/status" `
  -Headers $adminHeaders `
  -Body '{"accountStatus":"DELETED","reason":"违规账号逻辑删除"}'
```

预期：

```text
HTTP 400
accountStatus 只能是 NORMAL、DISABLED 或 BANNED
```

### 7.6 修改管理员账号状态应失败

```powershell
$adminUserId = 1

Invoke-WebRequest `
  -Method PUT `
  -Uri "$baseUrl/admin/users/$adminUserId/status" `
  -Headers $adminHeaders `
  -Body '{"accountStatus":"DISABLED","reason":"测试"}'
```

预期：

```text
HTTP 403
管理员账号不能通过该接口修改状态
```

## 八、管理员修改商品状态测试

### 8.1 下架商品

```powershell
$productId = 1
$productReportId = 1 # 替换为 target_type=PRODUCT,target_id=$productId 的举报 ID

Invoke-RestMethod `
  -Method PUT `
  -Uri "$baseUrl/admin/products/$productId/status" `
  -Headers $adminHeaders `
  -Body "{`"status`":`"OFF_SHELF`",`"reason`":`"举报成立，下架商品`",`"relatedType`":`"REPORT`",`"relatedId`":$productReportId}"
```

预期：

```text
HTTP 200
status = OFF_SHELF
off_shelf_reason 已记录
```

数据库检查：

```sql
SELECT id, status, off_shelf_reason, is_deleted
FROM products
WHERE id = 1;

SELECT id, operation_type, target_type, target_id, description, related_type, related_id
FROM admin_logs
WHERE operation_type = 'UPDATE_PRODUCT_STATUS'
ORDER BY id DESC;
```

### 8.2 恢复商品上架

```powershell
$productAppealId = 1 # 替换为 target_type=PRODUCT,target_id=$productId 的申诉 ID

Invoke-RestMethod `
  -Method PUT `
  -Uri "$baseUrl/admin/products/$productId/status" `
  -Headers $adminHeaders `
  -Body "{`"status`":`"ON_SALE`",`"reason`":`"申诉通过，恢复商品`",`"relatedType`":`"APPEAL`",`"relatedId`":$productAppealId}"
```

预期：

```text
HTTP 200
status = ON_SALE
off_shelf_reason 清空
is_deleted = 0
```

### 8.3 逻辑删除商品

```powershell
Invoke-RestMethod `
  -Method PUT `
  -Uri "$baseUrl/admin/products/$productId/status" `
  -Headers $adminHeaders `
  -Body '{"status":"DELETED","reason":"违规商品逻辑删除"}'
```

预期：

```text
HTTP 200
status = DELETED
is_deleted = 1
不会物理删除 products 记录
```

## 九、管理员修改订单状态测试

### 9.1 异常关闭订单

```powershell
$orderId = 1
$orderReportId = 1 # 替换为 target_type=ORDER,target_id=$orderId 的举报 ID

Invoke-RestMethod `
  -Method POST `
  -Uri "$baseUrl/admin/orders/$orderId/exception-close" `
  -Headers $adminHeaders `
  -Body "{`"reason`":`"举报成立，管理员异常关闭订单`",`"responsibleParty`":`"SELLER`",`"relatedType`":`"REPORT`",`"relatedId`":$orderReportId}"
```

预期：

```text
HTTP 200
status = EXCEPTION_CLOSED
cancel_reason 已记录
cancel_by = 当前管理员 ID
close_time 非空
responsibleParty = SELLER 时，关联商品 status = OFF_SHELF
```

数据库检查：

```sql
SELECT id, product_id, status, cancel_reason, cancel_by, close_time
FROM orders
WHERE id = 1;

SELECT p.id, p.status
FROM products p
JOIN orders o ON o.product_id = p.id
WHERE o.id = 1;

SELECT id, operation_type, target_type, target_id, description, related_type, related_id
FROM admin_logs
WHERE operation_type = 'ORDER_EXCEPTION_CLOSE'
ORDER BY id DESC;
```

### 9.2 完成订单

```powershell
Invoke-RestMethod `
  -Method PUT `
  -Uri "$baseUrl/admin/orders/$orderId/status" `
  -Headers $adminHeaders `
  -Body '{"status":"COMPLETED","reason":"管理员确认交易完成"}'
```

预期：

```text
HTTP 200
status = COMPLETED
finish_time 非空
关联商品 status = SOLD
```

### 9.3 取消订单

```powershell
Invoke-RestMethod `
  -Method PUT `
  -Uri "$baseUrl/admin/orders/$orderId/status" `
  -Headers $adminHeaders `
  -Body '{"status":"CANCELED","reason":"管理员取消订单"}'
```

预期：

```text
HTTP 200
status = CANCELED
cancel_reason 已记录
close_time 非空
关联商品 status = ON_SALE
```

### 9.4 设置无效订单状态应失败

```powershell
Invoke-WebRequest `
  -Method PUT `
  -Uri "$baseUrl/admin/orders/$orderId/status" `
  -Headers $adminHeaders `
  -Body '{"status":"DELETED","reason":"订单不支持该状态"}'
```

预期：

```text
HTTP 400
message = status must be PENDING_CONFIRM, WAIT_MEET, COMPLETED or CANCELED; use exception-close for EXCEPTION_CLOSED
```

## 十、超时订单清理说明

`POST /admin/orders/cleanup-expired` 管理员接口已移除，超时未确认订单由后端服务定时任务自动清理。

当前不再做该接口的权限测试。需要回归验证时，重点检查：

- 服务启动后每 1 分钟触发一次清理任务。
- 仅 `status='PENDING_CONFIRM'` 且 `expire_time < NOW()` 的订单会被取消。
- 只有数据库更新返回 `RowsAffected=1` 的订单才会解锁商品和发送超时通知。
- 并发执行或多实例部署时，同一订单不会重复发送超时取消消息。

## 十一、回归测试建议

### 11.1 举报和申诉最终状态

```sql
SELECT status, COUNT(*)
FROM reports
GROUP BY status;

SELECT status, COUNT(*)
FROM appeals
GROUP BY status;
```

预期状态集合：

```text
reports: PENDING / PROCESSING / APPROVED / REJECTED / CLOSED
appeals: PENDING / PROCESSING / APPROVED / REJECTED / CLOSED
```

### 11.2 确认没有物理删除

账号注销：

```sql
SELECT id, account_status, is_deleted
FROM users
WHERE account_status = 'CANCELED';
```

商品逻辑删除：

```sql
SELECT id, status, is_deleted
FROM products
WHERE status = 'DELETED';
```

预期：

```text
记录仍存在，只是状态或 is_deleted 变化
```

## 十二、后端编译测试

在 `backend` 目录执行：

```powershell
go test ./...
```

预期：

```text
全部 package 通过
```
