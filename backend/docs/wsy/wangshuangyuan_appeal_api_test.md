# 王双媛申诉模块接口测试说明

## 一、测试范围

本文档用于记录申诉模块接口测试过程与结果，覆盖用户账号、商品、订单、举报四类申诉对象。

用户端接口：

```http
POST /appeals
GET /appeals/my
GET /appeals/:id
POST /appeals/:id/close
```

管理员接口：

```http
GET /admin/appeals
GET /admin/appeals/:id
POST /admin/appeals/:id/processing
POST /admin/appeals/:id/handle
```

关联后续处置接口：

```http
PUT /admin/users/:id/status
PUT /admin/products/:id/status
POST /admin/orders/:id/exception-close
PUT /admin/orders/:id/status
```

测试重点：

- 用户可以提交 `PRODUCT`、`USER`、`ORDER`、`REPORT` 类型申诉。
- 管理员可以查看、标记处理中、通过或驳回申诉。
- `PENDING` 状态申诉允许用户撤回。
- `PROCESSING` 状态申诉不允许用户撤回，也不允许重复提交同一目标申诉。
- 申诉处理为 `APPROVED` 后，不自动恢复账号、商品、订单或举报状态。
- 后续业务处置必须调用对应后台接口，并通过 `relatedType=APPEAL`、`relatedId=申诉ID` 关联来源。
- 管理员处理申诉会写入 `admin_logs`，并给申诉人发送站内消息。

## 二、前置条件

后端服务默认地址：

```text
http://localhost:8080
```

启动后端：

```powershell
cd D:\cau-used-goods-app\backend
go run ./cmd/server
```

数据库需已导入最新表结构：

```text
backend/scripts/sql/schema.sql
```

相关数据表：

```text
users
products
orders
reports
appeals
appeal_images
messages
admin_logs
```

申诉对象类型：

```text
PRODUCT
USER
ORDER
REPORT
```

申诉状态：

```text
PENDING
PROCESSING
APPROVED
REJECTED
CLOSED
```

状态流转约定：

```text
用户提交申诉: PENDING
用户主动撤回: PENDING -> CLOSED
管理员接手处理: PENDING -> PROCESSING
管理员通过申诉: PENDING/PROCESSING -> APPROVED
管理员驳回申诉: PENDING/PROCESSING -> REJECTED
```

## 三、公共变量和登录

### 3.1 设置基础地址

```powershell
$baseUrl = "http://localhost:8080"
```

### 3.2 登录管理员

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

### 3.3 登录普通用户

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

### 3.4 初始化用户状态

进入 MySQL：

```powershell
mysql -h 127.0.0.1 -P 3306 -u root -p --protocol=tcp cau_used_goods
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
    role = 'ADMIN',
    nickname = '申诉测试管理员'
WHERE openid = 'dev_appeal_admin_001';
```

## 四、账号申诉流程测试

### 4.1 模拟账号被禁用

```sql
UPDATE users
SET account_status = 'DISABLED'
WHERE openid = 'dev_appeal_user_001';
```

### 4.2 用户提交账号申诉

```powershell
$userAppeal = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/appeals" `
  -Headers $userHeaders `
  -ContentType "application/json" `
  -Body "{`"targetType`":`"USER`",`"targetId`":$userId,`"reason`":`"账号被误禁用，申请恢复正常使用`",`"evidenceUrls`":[`"/uploads/appeal/user-demo.jpg`"]}"

$userAppealId = $userAppeal.data.id
$userAppeal.data
```

预期结果：

- 返回 `code = 0`。
- `targetType = USER`。
- `targetId = $userId`。
- `status = PENDING`。

### 4.3 管理员标记为处理中

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/admin/appeals/$userAppealId/processing" `
  -Headers $adminHeaders
```

预期结果：

- `status = PROCESSING`。
- `admin_logs` 写入 `MARK_APPEAL_PROCESSING`。

### 4.4 PROCESSING 后用户不能撤回

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/appeals/$userAppealId/close" `
    -Headers $userHeaders `
    -ContentType "application/json" `
    -Body '{"closeReason":"用户尝试在处理中状态撤回申诉"}'
} catch {
  $_.Exception.Response.StatusCode.value__
  $_.ErrorDetails.Message
}
```

预期结果：

```text
400
message = appeal cannot be closed
```

### 4.5 管理员通过账号申诉

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/admin/appeals/$userAppealId/handle" `
  -Headers $adminHeaders `
  -ContentType "application/json" `
  -Body '{"status":"APPROVED","handleResult":"申诉通过，确认账号可恢复正常使用"}'
```

预期结果：

- `status = APPROVED`。
- `handlerId = $adminId`。
- `handleTime` 不为空。
- `admin_logs` 写入 `APPROVE_APPEAL`。
- `messages` 写入申诉处理结果通知。

### 4.6 账号不会自动恢复

```sql
SELECT id, openid, account_status
FROM users
WHERE id = 用户ID;
```

预期结果：

```text
account_status 仍为 DISABLED
```

### 4.7 管理员恢复账号状态

```powershell
Invoke-RestMethod `
  -Method Put `
  -Uri "$baseUrl/admin/users/$userId/status" `
  -Headers $adminHeaders `
  -ContentType "application/json" `
  -Body "{`"accountStatus`":`"NORMAL`",`"reason`":`"申诉通过，恢复账号正常状态`",`"relatedType`":`"APPEAL`",`"relatedId`":$userAppealId}"
```

预期结果：

- 用户 `account_status = NORMAL`。
- 后续处置日志关联 `related_type = APPEAL`、`related_id = $userAppealId`。

## 五、商品申诉流程测试

### 5.1 准备下架商品

```sql
UPDATE users
SET account_status = 'NORMAL',
    auth_status = 'VERIFIED'
WHERE openid = 'dev_appeal_user_001';

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
  '商品申诉测试商品',
  '用于测试商品被下架后提交申诉并恢复上架',
  30.00,
  '九成新',
  '学校西门',
  'OFF_SHELF',
  '测试下架'
);

SELECT LAST_INSERT_ID() AS product_id;
```

PowerShell 设置商品 ID：

```powershell
$productId = 这里填商品ID
```

### 5.2 用户提交商品申诉

```powershell
$productAppeal = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/appeals" `
  -Headers $userHeaders `
  -ContentType "application/json" `
  -Body "{`"targetType`":`"PRODUCT`",`"targetId`":$productId,`"reason`":`"商品被误判违规下架，申请恢复上架`",`"evidenceUrls`":[`"/uploads/appeal/product-demo.jpg`"]}"

$productAppealId = $productAppeal.data.id
$productAppeal.data
```

预期结果：

- `targetType = PRODUCT`。
- `targetId = $productId`。
- `status = PENDING`。

### 5.3 管理员查看并标记处理中

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "$baseUrl/admin/appeals?targetType=PRODUCT&status=PENDING&page=1&pageSize=20" `
  -Headers $adminHeaders

Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/admin/appeals/$productAppealId/processing" `
  -Headers $adminHeaders
```

预期结果：

- 管理员可查询到商品申诉。
- 标记后 `status = PROCESSING`。

### 5.4 PROCESSING 后不能重复提交同一商品申诉

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/appeals" `
    -Headers $userHeaders `
    -ContentType "application/json" `
    -Body "{`"targetType`":`"PRODUCT`",`"targetId`":$productId,`"reason`":`"重复商品申诉测试`"}"
} catch {
  $_.Exception.Response.StatusCode.value__
  $_.ErrorDetails.Message
}
```

预期结果：

```text
409
message = you already have an active appeal for this target
```

### 5.5 管理员通过商品申诉

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/admin/appeals/$productAppealId/handle" `
  -Headers $adminHeaders `
  -ContentType "application/json" `
  -Body '{"status":"APPROVED","handleResult":"申诉通过，确认商品可以恢复上架"}'
```

预期结果：

- `status = APPROVED`。
- 写入 `APPROVE_APPEAL` 日志。
- 给申诉人发送站内消息。

### 5.6 商品不会自动恢复上架

```sql
SELECT id, title, status, off_shelf_reason
FROM products
WHERE id = 商品ID;
```

预期结果：

```text
status 仍为 OFF_SHELF
```

### 5.7 管理员恢复商品上架

```powershell
Invoke-RestMethod `
  -Method Put `
  -Uri "$baseUrl/admin/products/$productId/status" `
  -Headers $adminHeaders `
  -ContentType "application/json" `
  -Body "{`"status`":`"ON_SALE`",`"reason`":`"申诉通过，恢复商品上架`",`"relatedType`":`"APPEAL`",`"relatedId`":$productAppealId}"
```

预期结果：

- 商品 `status = ON_SALE`。
- 后续处置日志关联该申诉。

## 六、订单申诉流程测试

### 6.1 登录买家和卖家

```powershell
$buyerLogin = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/auth/dev-login" `
  -ContentType "application/json" `
  -Body '{"openid":"dev_appeal_order_buyer_001","role":"USER"}'

$buyerToken = $buyerLogin.data.token
$buyerId = $buyerLogin.data.user.id
$buyerHeaders = @{ Authorization = "Bearer $buyerToken" }

$sellerLogin = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/auth/dev-login" `
  -ContentType "application/json" `
  -Body '{"openid":"dev_appeal_order_seller_001","role":"USER"}'

$sellerToken = $sellerLogin.data.token
$sellerId = $sellerLogin.data.user.id
$sellerHeaders = @{ Authorization = "Bearer $sellerToken" }
```

### 6.2 准备订单

```sql
UPDATE users
SET auth_status = 'VERIFIED',
    account_status = 'NORMAL',
    nickname = '订单申诉测试买家'
WHERE openid = 'dev_appeal_order_buyer_001';

UPDATE users
SET auth_status = 'VERIFIED',
    account_status = 'NORMAL',
    nickname = '订单申诉测试卖家'
WHERE openid = 'dev_appeal_order_seller_001';

INSERT INTO products (
  seller_id,
  category_id,
  title,
  description,
  price,
  condition_level,
  meet_location,
  status
) VALUES (
  (SELECT id FROM users WHERE openid = 'dev_appeal_order_seller_001' LIMIT 1),
  (SELECT id FROM categories ORDER BY id ASC LIMIT 1),
  '订单申诉测试商品',
  '用于测试订单申诉处理流程',
  50.00,
  '九成新',
  '学校西门',
  'LOCKED'
);

SET @product_id = LAST_INSERT_ID();

INSERT INTO orders (
  order_no,
  product_id,
  buyer_id,
  seller_id,
  product_title_snapshot,
  product_price_snapshot,
  status,
  remark,
  meet_location,
  expire_time
) VALUES (
  CONCAT('APPEALORDER', DATE_FORMAT(NOW(), '%Y%m%d%H%i%s')),
  @product_id,
  (SELECT id FROM users WHERE openid = 'dev_appeal_order_buyer_001' LIMIT 1),
  (SELECT id FROM users WHERE openid = 'dev_appeal_order_seller_001' LIMIT 1),
  '订单申诉测试商品',
  50.00,
  'WAIT_MEET',
  '订单申诉测试订单',
  '学校西门',
  DATE_ADD(NOW(), INTERVAL 1 DAY)
);

SELECT @product_id AS product_id, LAST_INSERT_ID() AS order_id;
```

PowerShell 设置订单 ID：

```powershell
$orderId = 这里填订单ID
```

### 6.3 买家提交订单申诉

```powershell
$orderAppeal = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/appeals" `
  -Headers $buyerHeaders `
  -ContentType "application/json" `
  -Body "{`"targetType`":`"ORDER`",`"targetId`":$orderId,`"reason`":`"订单处理存在争议，申请管理员复核`",`"evidenceUrls`":[`"/uploads/appeal/order-demo.jpg`"]}"

$orderAppealId = $orderAppeal.data.id
$orderAppeal.data
```

预期结果：

- `targetType = ORDER`。
- `targetId = $orderId`。
- `status = PENDING`。

### 6.4 管理员处理订单申诉

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/admin/appeals/$orderAppealId/processing" `
  -Headers $adminHeaders

Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/admin/appeals/$orderAppealId/handle" `
  -Headers $adminHeaders `
  -ContentType "application/json" `
  -Body '{"status":"APPROVED","handleResult":"申诉通过，订单后续由管理员按实际情况处理"}'
```

预期结果：

- 申诉 `status = APPROVED`。
- 订单状态不会自动变化。

### 6.5 管理员异常关闭订单

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/admin/orders/$orderId/exception-close" `
  -Headers $adminHeaders `
  -ContentType "application/json" `
  -Body "{`"reason`":`"申诉通过，管理员异常关闭订单`",`"responsibleParty`":`"SELLER`",`"relatedType`":`"APPEAL`",`"relatedId`":$orderAppealId}"
```

预期结果：

- 订单 `status = EXCEPTION_CLOSED`。
- 订单处理日志关联 `related_type = APPEAL`、`related_id = $orderAppealId`。

## 七、举报申诉流程测试

### 7.1 准备举报记录

```sql
INSERT INTO reports (
  reporter_id,
  target_type,
  target_id,
  reason_type,
  description,
  status
) VALUES (
  (SELECT id FROM users WHERE openid = 'dev_appeal_user_001' LIMIT 1),
  'USER',
  (SELECT id FROM users WHERE openid = 'dev_appeal_admin_001' LIMIT 1),
  'OTHER',
  '用于测试 REPORT 类型申诉',
  'REJECTED'
);

SELECT LAST_INSERT_ID() AS report_id;
```

PowerShell 设置举报 ID：

```powershell
$reportId = 这里填举报ID
```

### 7.2 举报人提交 REPORT 申诉

```powershell
$reportAppeal = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/appeals" `
  -Headers $userHeaders `
  -ContentType "application/json" `
  -Body "{`"targetType`":`"REPORT`",`"targetId`":$reportId,`"reason`":`"对举报处理结果有异议，申请复核`"}"

$reportAppealId = $reportAppeal.data.id
$reportAppeal.data
```

预期结果：

- `targetType = REPORT`。
- `targetId = $reportId`。
- `status = PENDING`。

### 7.3 管理员驳回 REPORT 申诉

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/admin/appeals/$reportAppealId/processing" `
  -Headers $adminHeaders

Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/admin/appeals/$reportAppealId/handle" `
  -Headers $adminHeaders `
  -ContentType "application/json" `
  -Body '{"status":"REJECTED","handleResult":"经复核，原举报处理结果无误，申诉驳回"}'
```

预期结果：

- 申诉 `status = REJECTED`。
- 写入 `REJECT_APPEAL` 管理员日志。
- 给申诉人发送站内消息。

## 八、用户主动撤回 PENDING 申诉

### 8.1 准备一条可撤回申诉

```powershell
$closeAppeal = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/appeals" `
  -Headers $userHeaders `
  -ContentType "application/json" `
  -Body "{`"targetType`":`"USER`",`"targetId`":$userId,`"reason`":`"用于测试用户主动撤回 PENDING 申诉`"}"

$closeAppealId = $closeAppeal.data.id
```

### 8.2 用户撤回申诉

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/appeals/$closeAppealId/close" `
  -Headers $userHeaders `
  -ContentType "application/json" `
  -Body '{"closeReason":"用户主动撤回申诉"}'
```

预期结果：

- 申诉 `status = CLOSED`。
- `handleResult` 保存撤回原因。

## 九、权限测试

### 9.1 未登录提交申诉

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/appeals" `
    -ContentType "application/json" `
    -Body '{"targetType":"USER","targetId":1,"reason":"未登录测试"}'
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期结果：

```text
401
```

### 9.2 普通用户访问后台申诉列表

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

预期结果：

```text
403
```

### 9.3 非本人查看申诉详情

```powershell
$otherLogin = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/auth/dev-login" `
  -ContentType "application/json" `
  -Body '{"openid":"dev_appeal_other_001","role":"USER"}'

$otherToken = $otherLogin.data.token
$otherHeaders = @{ Authorization = "Bearer $otherToken" }

try {
  Invoke-RestMethod `
    -Method Get `
    -Uri "$baseUrl/appeals/$productAppealId" `
    -Headers $otherHeaders
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期结果：

```text
403
```

### 9.4 非相关用户提交订单申诉

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/appeals" `
    -Headers $otherHeaders `
    -ContentType "application/json" `
    -Body "{`"targetType`":`"ORDER`",`"targetId`":$orderId,`"reason`":`"非订单相关用户提交申诉`"}"
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期结果：

```text
403
```

## 十、异常参数测试

### 10.1 非法 targetType

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
  $_.ErrorDetails.Message
}
```

预期结果：

```text
400
```

### 10.2 不存在的申诉对象

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
  $_.ErrorDetails.Message
}
```

预期结果：

```text
404
message = appeal target not found
```

### 10.3 缺少 reason

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/appeals" `
    -Headers $userHeaders `
    -ContentType "application/json" `
    -Body '{"targetType":"USER","targetId":1}'
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期结果：

```text
400
```

### 10.4 管理员处理状态非法

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/admin/appeals/$userAppealId/handle" `
    -Headers $adminHeaders `
    -ContentType "application/json" `
    -Body '{"status":"CLOSED","handleResult":"管理员不能直接处理为关闭"}'
} catch {
  $_.Exception.Response.StatusCode.value__
  $_.ErrorDetails.Message
}
```

预期结果：

```text
400
message = status must be APPROVED or REJECTED
```

### 10.5 重复处理已完结申诉

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/admin/appeals/$userAppealId/handle" `
    -Headers $adminHeaders `
    -ContentType "application/json" `
    -Body '{"status":"REJECTED","handleResult":"重复处理测试"}'
} catch {
  $_.Exception.Response.StatusCode.value__
  $_.ErrorDetails.Message
}
```

预期结果：

```text
409
message = appeal already handled
```

## 十一、数据库核验

### 11.1 查询申诉

```sql
SELECT id, appellant_id, target_type, target_id, status, handler_id, handle_result, handle_time
FROM appeals
ORDER BY id DESC;
```

### 11.2 查询管理员日志

```sql
SELECT admin_id, operation_type, target_type, target_id, related_type, related_id, description, create_time
FROM admin_logs
WHERE target_type = 'APPEAL'
   OR related_type = 'APPEAL'
ORDER BY id DESC;
```

预期包含：

```text
MARK_APPEAL_PROCESSING
APPROVE_APPEAL
REJECT_APPEAL
```

后续处置日志应关联：

```text
related_type = APPEAL
related_id = 对应申诉ID
```

### 11.3 查询站内消息

```sql
SELECT receiver_id, sender_id, message_type, title, content, related_type, related_id, read_status, create_time
FROM messages
WHERE related_type = 'APPEAL'
ORDER BY id DESC;
```

预期结果：

- `receiver_id` 为申诉人 ID。
- `sender_id` 为管理员 ID。
- `message_type = SYSTEM_NOTICE`。
- `related_type = APPEAL`。
- `related_id` 为对应申诉 ID。
- `read_status = UNREAD`。

## 十二、测试结论

测试模块：申诉模块

测试接口：

```text
POST /appeals
GET /appeals/my
GET /appeals/:id
POST /appeals/:id/close
GET /admin/appeals
GET /admin/appeals/:id
POST /admin/appeals/:id/processing
POST /admin/appeals/:id/handle
```

测试结果：

- 用户可以提交 `USER`、`PRODUCT`、`ORDER`、`REPORT` 类型申诉。
- 用户只能查看自己的申诉详情。
- 用户可以撤回 `PENDING` 状态申诉。
- 管理员可以查看全部申诉并查看详情。
- 管理员可以将 `PENDING` 申诉标记为 `PROCESSING`。
- 申诉进入 `PROCESSING` 后，用户不能撤回，也不能重复提交同一目标申诉。
- 管理员可以将申诉处理为 `APPROVED` 或 `REJECTED`。
- 管理员不能将申诉直接处理为 `CLOSED`。
- 已完结申诉不能重复处理。
- 申诉通过后，不会自动恢复账号、商品、订单或举报状态。
- 账号、商品、订单的后续处置通过对应后台接口执行，并关联 `APPEAL` 来源。
- 管理员处理申诉后写入 `admin_logs`。
- 管理员处理申诉后给申诉人发送站内消息。
- 权限校验和非法参数处理正常。

结论：

```text
申诉模块接口测试通过。
```

## 十三、提交建议

查看改动：

```powershell
git status
git diff -- backend/docs/wsy/wangshuangyuan_appeal_api_test.md
```

提交文档：

```powershell
git add backend/docs/wsy/wangshuangyuan_appeal_api_test.md
git commit -m "docs(appeal): 完善申诉模块接口测试文档"
```
