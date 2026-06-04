# 王双媛申诉模块接口测试说明

## 一、测试范围

本说明用于测试申诉模块接口：

```http
POST /appeals
GET /appeals/my
GET /appeals/:id
GET /admin/appeals
GET /admin/appeals/:id
POST /admin/appeals/:id/handle
```

模块代码位置：

```text
backend/internal/appeal/
backend/cmd/server/main.go
```

关联数据表：

```text
appeals
appeal_images
users
products
orders
reports
messages
admin_logs
```

## 二、测试前置条件

本地服务默认地址：

```text
http://localhost:8080
```

确认配置：

```yaml
server:
  env: dev
  port: 8080
```

确认数据库已执行：

```text
backend/scripts/sql/schema.sql
```

确认存在表：

```sql
SHOW TABLES LIKE 'appeal%';
```

预期：

```text
appeals
appeal_images
```

## 三、准备测试用户

### 3.1 设置基础变量

```powershell
$baseUrl = "http://localhost:8080"
```

### 3.2 登录普通用户

```powershell
$userLogin = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/auth/dev-login" `
  -ContentType "application/json" `
  -Body '{"openid":"dev_appeal_user_001","role":"USER"}'

$userToken = $userLogin.data.token
$userId = $userLogin.data.user.id
$userHeaders = @{ Authorization = "Bearer $userToken" }

$userId
```

### 3.3 登录管理员

```powershell
$adminLogin = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/auth/dev-login" `
  -ContentType "application/json" `
  -Body '{"openid":"dev_appeal_admin_001","role":"ADMIN"}'

$adminToken = $adminLogin.data.token
$adminId = $adminLogin.data.user.id
$adminHeaders = @{ Authorization = "Bearer $adminToken" }

$adminId
```

### 3.4 更新测试用户状态

进入 MySQL：

```powershell
mysql -h 127.0.0.1 -P 3306 -u root -p cau_used_goods
```

执行：

```sql
UPDATE users
SET auth_status = 'VERIFIED',
    account_status = 'NORMAL',
    nickname = '申诉测试用户'
WHERE openid = 'dev_appeal_user_001';

UPDATE users
SET auth_status = 'VERIFIED',
    account_status = 'NORMAL',
    nickname = '申诉测试管理员',
    role = 'ADMIN'
WHERE openid = 'dev_appeal_admin_001';
```

## 四、准备商品申诉数据

创建一个被下架商品，用于测试申诉通过后恢复上架。

```sql
INSERT INTO products (
  seller_id,
  category_id,
  title,
  description,
  price,
  condition_level,
  meet_location,
  status,
  off_shelf_reason
) VALUES (
  (SELECT id FROM users WHERE openid = 'dev_appeal_user_001' LIMIT 1),
  (SELECT id FROM categories ORDER BY id ASC LIMIT 1),
  '申诉测试商品',
  '用于测试商品申诉通过后恢复上架',
  30.00,
  '九成新',
  '学校西门',
  'OFF_SHELF',
  '测试下架'
);

SET @appeal_product_id = LAST_INSERT_ID();

SELECT @appeal_product_id AS product_id;
```

记录输出的 `product_id`。

在 PowerShell 设置：

```powershell
$productId = 这里填写商品ID
```

## 五、正常流程测试

### 5.1 用户提交商品申诉

```powershell
$appeal = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/appeals" `
  -Headers $userHeaders `
  -ContentType "application/json" `
  -Body "{`"targetType`":`"PRODUCT`",`"targetId`":$productId,`"reason`":`"商品被误判违规，申请恢复上架`",`"evidenceUrls`":[`"/uploads/appeal/demo.jpg`"]}"

$appeal.data
$appealId = $appeal.data.id
$appealId
```

预期：

- 返回申诉 ID。
- `targetType = PRODUCT`。
- `targetId = $productId`。
- `status = PENDING`。
- `images` 包含提交的凭证路径。
- 商品申诉要求当前用户是该商品卖家。

### 5.2 重复提交未处理申诉

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/appeals" `
    -Headers $userHeaders `
    -ContentType "application/json" `
    -Body "{`"targetType`":`"PRODUCT`",`"targetId`":$productId,`"reason`":`"重复申诉测试`"}"
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期：

```text
409
```

响应消息：

```text
you already have an active appeal for this target
```

### 5.3 用户查看自己的申诉列表

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "$baseUrl/appeals/my?page=1&pageSize=20" `
  -Headers $userHeaders
```

预期：

- 返回当前用户自己的申诉。
- 包含刚创建的商品申诉。

### 5.4 用户查看申诉详情

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "$baseUrl/appeals/$appealId" `
  -Headers $userHeaders
```

预期：

- 返回申诉详情。
- `appellantId` 等于当前用户 ID。

### 5.5 管理员查看申诉列表

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "$baseUrl/admin/appeals?status=PENDING&page=1&pageSize=20" `
  -Headers $adminHeaders
```

预期：

- 管理员可以查看待处理申诉。
- 包含申诉人昵称。

### 5.6 管理员查看申诉详情

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "$baseUrl/admin/appeals/$appealId" `
  -Headers $adminHeaders
```

预期：

- 管理员可以查看完整申诉详情。

### 5.7 管理员通过商品申诉

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/admin/appeals/$appealId/handle" `
  -Headers $adminHeaders `
  -ContentType "application/json" `
  -Body '{"status":"APPROVED","handleResult":"申诉通过，商品恢复上架"}'
```

预期：

- `status = APPROVED`。
- `handlerId` 等于管理员 ID。
- `handleResult` 为处理说明。
- `handleTime` 不为空。

### 5.8 检查商品是否恢复上架

进入 MySQL 执行：

```sql
SELECT id, status, off_shelf_reason
FROM products
WHERE id = 商品ID;
```

预期：

```text
status = ON_SALE
off_shelf_reason = NULL
```

### 5.9 检查站内消息

```sql
SELECT receiver_id, sender_id, message_type, title, content, related_type, related_id, read_status
FROM messages
WHERE related_type = 'APPEAL' AND related_id = 申诉ID;
```

预期：

- `receiver_id` 是申诉人 ID。
- `sender_id` 是管理员 ID。
- `message_type = SYSTEM_NOTICE`。
- `related_type = APPEAL`。
- `read_status = UNREAD`。

### 5.10 检查管理员日志

```sql
SELECT admin_id, operation_type, target_type, target_id, description
FROM admin_logs
WHERE target_type = 'APPEAL' AND target_id = 申诉ID;
```

预期：

- `admin_id` 是管理员 ID。
- `operation_type = HANDLE_APPEAL`。
- `target_type = APPEAL`。
- `target_id` 是申诉 ID。

## 六、账号申诉联动测试

### 6.1 准备被禁用用户

```powershell
$disabledLogin = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/auth/dev-login" `
  -ContentType "application/json" `
  -Body '{"openid":"dev_appeal_disabled_user_001","role":"USER"}'

$disabledToken = $disabledLogin.data.token
$disabledId = $disabledLogin.data.user.id
$disabledHeaders = @{ Authorization = "Bearer $disabledToken" }
$disabledId
```

MySQL 中执行：

```sql
UPDATE users
SET auth_status = 'VERIFIED',
    account_status = 'DISABLED',
    nickname = '申诉测试禁用用户'
WHERE openid = 'dev_appeal_disabled_user_001';
```

### 6.2 禁用用户提交账号申诉

```powershell
$userAppeal = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/appeals" `
  -Headers $disabledHeaders `
  -ContentType "application/json" `
  -Body "{`"targetType`":`"USER`",`"targetId`":$disabledId,`"reason`":`"账号被误封，申请恢复`"}"

$userAppealId = $userAppeal.data.id
$userAppealId
```

预期：

- 即使账号状态为 `DISABLED`，仍可以提交申诉。
- 返回 `status = PENDING`。

### 6.3 管理员通过账号申诉

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/admin/appeals/$userAppealId/handle" `
  -Headers $adminHeaders `
  -ContentType "application/json" `
  -Body '{"status":"APPROVED","handleResult":"申诉通过，账号恢复正常"}'
```

检查账号状态：

```sql
SELECT id, account_status
FROM users
WHERE id = 禁用用户ID;
```

预期：

```text
account_status = NORMAL
```

## 七、异常场景测试

### 7.1 未登录提交申诉

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/appeals" `
    -ContentType "application/json" `
    -Body '{"targetType":"PRODUCT","targetId":1,"reason":"未登录测试"}'
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期：

```text
401
```

### 7.2 非管理员访问后台申诉列表

```powershell
try {
  Invoke-RestMethod `
    -Method Get `
    -Uri "$baseUrl/admin/appeals" `
    -Headers $userHeaders
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期：

```text
403
```

### 7.3 非本人查看申诉详情

登录第二个普通用户：

```powershell
$otherLogin = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/auth/dev-login" `
  -ContentType "application/json" `
  -Body '{"openid":"dev_appeal_other_001","role":"USER"}'

$otherHeaders = @{ Authorization = "Bearer $($otherLogin.data.token)" }
```

访问第一个用户的申诉：

```powershell
try {
  Invoke-RestMethod `
    -Method Get `
    -Uri "$baseUrl/appeals/$appealId" `
    -Headers $otherHeaders
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期：

```text
403
```

### 7.4 不存在的申诉对象

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/appeals" `
    -Headers $userHeaders `
    -ContentType "application/json" `
    -Body '{"targetType":"PRODUCT","targetId":999999,"reason":"不存在对象测试"}'
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期：

```text
404
```

响应消息：

```text
appeal target not found
```

### 7.5 非法申诉类型

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/appeals" `
    -Headers $userHeaders `
    -ContentType "application/json" `
    -Body '{"targetType":"INVALID","targetId":1,"reason":"非法类型测试"}'
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期：

```text
400
```

### 7.6 重复处理申诉

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/admin/appeals/$appealId/handle" `
    -Headers $adminHeaders `
    -ContentType "application/json" `
    -Body '{"status":"REJECTED","handleResult":"重复处理测试"}'
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期：

```text
409
```

响应消息：

```text
appeal already handled
```

## 八、Apifox 测试建议

建议新建分组：

```text
王双媛-申诉模块
```

环境变量：

```text
baseUrl = http://localhost:8080
userToken = 普通用户 token
adminToken = 管理员 token
productId = 测试商品 ID
appealId = 申诉 ID
```

接口列表：

```text
POST {{baseUrl}}/appeals
GET  {{baseUrl}}/appeals/my?page=1&pageSize=20
GET  {{baseUrl}}/appeals/{{appealId}}
GET  {{baseUrl}}/admin/appeals?page=1&pageSize=20
GET  {{baseUrl}}/admin/appeals/{{appealId}}
POST {{baseUrl}}/admin/appeals/{{appealId}}/handle
```

## 九、测试结论模板

```text
测试模块：申诉模块

测试接口：
- POST /appeals
- GET /appeals/my
- GET /appeals/:id
- GET /admin/appeals
- GET /admin/appeals/:id
- POST /admin/appeals/:id/handle

测试结果：
- 用户提交申诉正常
- 重复提交未处理申诉被拦截
- 用户只能查看自己的申诉
- 管理员可以查看全部申诉
- 管理员处理申诉正常
- 商品申诉通过后恢复上架
- 账号申诉通过后恢复 NORMAL
- 处理后生成站内消息
- 处理后写入管理员日志
- 非法参数和权限拦截正常

结论：申诉模块接口测试通过
```
