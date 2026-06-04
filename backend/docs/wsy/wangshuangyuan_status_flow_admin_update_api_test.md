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
POST /admin/orders/cleanup-expired
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

Invoke-RestMethod `
  -Method PUT `
  -Uri "$baseUrl/admin/users/$userId/status" `
  -Headers $adminHeaders `
  -Body '{"accountStatus":"DISABLED","reason":"举报成立，禁用账号"}'
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

SELECT id, operation_type, target_type, target_id, description
FROM admin_logs
WHERE operation_type = 'UPDATE_USER_STATUS'
ORDER BY id DESC;
```

### 7.2 恢复账号

```powershell
Invoke-RestMethod `
  -Method PUT `
  -Uri "$baseUrl/admin/users/$userId/status" `
  -Headers $adminHeaders `
  -Body '{"accountStatus":"NORMAL","reason":"申诉通过，恢复账号"}'
```

预期：

```text
HTTP 200
accountStatus = NORMAL
is_deleted = 0
```

### 7.3 逻辑删除账号

```powershell
Invoke-RestMethod `
  -Method PUT `
  -Uri "$baseUrl/admin/users/$userId/status" `
  -Headers $adminHeaders `
  -Body '{"accountStatus":"DELETED","reason":"违规账号逻辑删除"}'
```

预期：

```text
HTTP 200
accountStatus = DELETED
is_deleted = 1
不会物理删除 users 记录
```

### 7.4 修改管理员账号状态应失败

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

Invoke-RestMethod `
  -Method PUT `
  -Uri "$baseUrl/admin/products/$productId/status" `
  -Headers $adminHeaders `
  -Body '{"status":"OFF_SHELF","reason":"举报成立，下架商品"}'
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

SELECT id, operation_type, target_type, target_id, description
FROM admin_logs
WHERE operation_type = 'UPDATE_PRODUCT_STATUS'
ORDER BY id DESC;
```

### 8.2 恢复商品上架

```powershell
Invoke-RestMethod `
  -Method PUT `
  -Uri "$baseUrl/admin/products/$productId/status" `
  -Headers $adminHeaders `
  -Body '{"status":"ON_SALE","reason":"申诉通过，恢复商品"}'
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

Invoke-RestMethod `
  -Method PUT `
  -Uri "$baseUrl/admin/orders/$orderId/status" `
  -Headers $adminHeaders `
  -Body '{"status":"EXCEPTION_CLOSED","reason":"举报成立，管理员异常关闭订单"}'
```

预期：

```text
HTTP 200
status = EXCEPTION_CLOSED
cancel_reason 已记录
cancel_by = 当前管理员 ID
close_time 非空
关联商品 status = ON_SALE
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

SELECT id, operation_type, target_type, target_id, description
FROM admin_logs
WHERE operation_type = 'UPDATE_ORDER_STATUS'
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
message = status must be PENDING_CONFIRM, WAIT_MEET, COMPLETED, CANCELED or EXCEPTION_CLOSED
```

## 十、管理员清理超时订单权限测试

### 10.1 管理员调用

```powershell
Invoke-RestMethod `
  -Method POST `
  -Uri "$baseUrl/admin/orders/cleanup-expired" `
  -Headers $adminHeaders
```

预期：

```text
HTTP 200
返回 cancelledCount
```

### 10.2 普通用户调用应失败

```powershell
Invoke-WebRequest `
  -Method POST `
  -Uri "$baseUrl/admin/orders/cleanup-expired" `
  -Headers $userHeaders
```

预期：

```text
HTTP 403
普通用户不能调用管理员清理接口
```

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

账号逻辑删除：

```sql
SELECT id, account_status, is_deleted
FROM users
WHERE account_status = 'DELETED';
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
