# 王双媛聊天模块接口测试说明

## 一、测试范围

本说明用于测试聊天模块接口：

```http
POST /chat/conversations
GET /chat/conversations
GET /chat/conversations/:id/messages
POST /chat/conversations/:id/messages
PUT /chat/conversations/:id/read
```

模块代码位置：

```text
backend/internal/chat/
backend/cmd/server/main.go
```

关联数据表：

```text
chat_conversations
chat_messages
products
users
product_images
```

## 二、测试前置条件

本地服务默认地址：

```text
http://localhost:8080
```

确认 `backend/config/config.yaml`：

```yaml
server:
  env: dev
  port: 8080
```

数据库已执行：

```text
backend/scripts/sql/schema.sql
```

需要确认数据库中存在：

```text
chat_conversations
chat_messages
```

聊天接口使用学生认证中间件，测试用户必须满足：

```text
auth_status = VERIFIED
account_status = NORMAL
```

## 三、准备测试用户和商品

### 3.1 启动后端

```powershell
cd D:\cau-used-goods-app\backend
go run .\cmd\server
```

以下命令在另一个 PowerShell 窗口执行。

### 3.2 设置基础变量

```powershell
$baseUrl = "http://localhost:8080"
```

### 3.3 登录买家用户

```powershell
$buyerLogin = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/auth/dev-login" `
  -ContentType "application/json" `
  -Body '{"openid":"dev_chat_buyer_001","role":"USER"}'

$buyerToken = $buyerLogin.data.token
$buyerId = $buyerLogin.data.user.id
$buyerHeaders = @{ Authorization = "Bearer $buyerToken" }

$buyerId
```

### 3.4 登录卖家用户

```powershell
$sellerLogin = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/auth/dev-login" `
  -ContentType "application/json" `
  -Body '{"openid":"dev_chat_seller_001","role":"USER"}'

$sellerToken = $sellerLogin.data.token
$sellerId = $sellerLogin.data.user.id
$sellerHeaders = @{ Authorization = "Bearer $sellerToken" }

$sellerId
```

### 3.5 将测试用户设为已认证

进入 MySQL：

```powershell
mysql -h 127.0.0.1 -P 3306 -u root -p cau_used_goods
```

执行 SQL，把下面的 `买家ID`、`卖家ID` 替换成上面输出的 `$buyerId`、`$sellerId`：

```sql
UPDATE users
SET auth_status = 'VERIFIED',
    account_status = 'NORMAL',
    nickname = '聊天测试买家'
WHERE id = 16;

UPDATE users
SET auth_status = 'VERIFIED',
    account_status = 'NORMAL',
    nickname = '聊天测试卖家'
WHERE id = 17;
```

### 3.6 准备测试商品

继续在 MySQL 中执行，把 `卖家ID` 替换成 `$sellerId`。

如果 `categories` 中已经有数据，可以直接使用第一个分类：

```sql
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
  卖家ID,
  (SELECT id FROM categories ORDER BY id ASC LIMIT 1),
  '聊天测试商品',
  '用于测试联系商家和聊天功能',
  20.00,
  '八成新',
  '学校东门',
  'ON_SALE'
);

SET @chat_product_id = LAST_INSERT_ID();

INSERT INTO product_images (product_id, image_url, sort_order)
VALUES (@chat_product_id, '/uploads/product/chat-test.jpg', 0);

SELECT @chat_product_id AS product_id;
```

记录输出的 `product_id`，后面记为：

```powershell
$productId = 这里填写商品ID
```

退出 MySQL：

```sql
exit;
```

在 PowerShell 中设置商品 ID：

```powershell
$productId = 1
```

注意把 `1` 改成刚才数据库输出的真实商品 ID。

## 四、正常流程测试

### 4.1 买家创建或获取聊天会话

```powershell
$conversation = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/chat/conversations" `
  -Headers $buyerHeaders `
  -ContentType "application/json" `
  -Body "{`"productId`":$productId}"

$conversation.data
$conversationId = $conversation.data.id
$conversationId
```

预期：

- 返回 `id`。
- `productId` 等于测试商品 ID。
- `buyerId` 等于买家 ID。
- `sellerId` 等于卖家 ID。
- `status = ACTIVE`。

### 4.2 重复创建会话

```powershell
$conversationAgain = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/chat/conversations" `
  -Headers $buyerHeaders `
  -ContentType "application/json" `
  -Body "{`"productId`":$productId}"

$conversationAgain.data.id
```

预期：

```text
返回的 id 与第一次的 conversationId 相同
```

说明唯一约束和复用会话逻辑正常。

### 4.3 买家发送消息

```powershell
$msg1 = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/chat/conversations/$conversationId/messages" `
  -Headers $buyerHeaders `
  -ContentType "application/json" `
  -Body '{"content":"你好，这个商品还在吗？"}'

$msg1.data
```

预期：

- `senderId` 等于买家 ID。
- `receiverId` 等于卖家 ID。
- `content` 为发送内容。
- `messageType = TEXT`。
- `readStatus = UNREAD`。

### 4.4 卖家查看会话列表

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "$baseUrl/chat/conversations?page=1&pageSize=20" `
  -Headers $sellerHeaders
```

预期：

- 能看到该会话。
- `targetUserId` 是买家 ID。
- `targetNickname` 是买家昵称。
- `lastMessageContent` 是买家刚发送的内容。
- `unreadCount` 大于等于 1。

### 4.5 卖家查看消息列表

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "$baseUrl/chat/conversations/$conversationId/messages?page=1&pageSize=30" `
  -Headers $sellerHeaders
```

预期：

- 返回买家发送的消息。
- 消息按 `create_time ASC, id ASC` 排序。

### 4.6 卖家标记已读

```powershell
Invoke-RestMethod `
  -Method Put `
  -Uri "$baseUrl/chat/conversations/$conversationId/read" `
  -Headers $sellerHeaders
```

预期：

```json
{
  "read": true,
  "count": 1
}
```

再次查询卖家会话列表：

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "$baseUrl/chat/conversations?page=1&pageSize=20" `
  -Headers $sellerHeaders
```

预期：

```text
unreadCount = 0
```

### 4.7 卖家回复消息

```powershell
$msg2 = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/chat/conversations/$conversationId/messages" `
  -Headers $sellerHeaders `
  -ContentType "application/json" `
  -Body '{"content":"还在，可以今天下午面交。"}'

$msg2.data
```

预期：

- `senderId` 等于卖家 ID。
- `receiverId` 等于买家 ID。
- `readStatus = UNREAD`。

### 4.8 买家查看消息并标记已读

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "$baseUrl/chat/conversations/$conversationId/messages?page=1&pageSize=30" `
  -Headers $buyerHeaders

Invoke-RestMethod `
  -Method Put `
  -Uri "$baseUrl/chat/conversations/$conversationId/read" `
  -Headers $buyerHeaders
```

预期：

- 买家能看到买家和卖家的完整消息。
- 买家标记已读后，买家的 `unreadCount` 变为 0。

## 五、异常场景测试

### 5.1 未登录访问

```powershell
try {
  Invoke-RestMethod `
    -Method Get `
    -Uri "$baseUrl/chat/conversations"
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期：

```text
401
```

### 5.2 未认证用户访问

新登录一个未认证用户：

```powershell
$unverifiedLogin = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/auth/dev-login" `
  -ContentType "application/json" `
  -Body '{"openid":"dev_chat_unverified_001","role":"USER"}'

$unverifiedHeaders = @{ Authorization = "Bearer $($unverifiedLogin.data.token)" }
```

调用聊天接口：

```powershell
try {
  Invoke-RestMethod `
    -Method Get `
    -Uri "$baseUrl/chat/conversations" `
    -Headers $unverifiedHeaders
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期：

```text
403
```

### 5.3 卖家不能联系自己的商品

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/chat/conversations" `
    -Headers $sellerHeaders `
    -ContentType "application/json" `
    -Body "{`"productId`":$productId}"
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期：

```text
400
```

响应消息：

```text
cannot chat with yourself
```

### 5.4 商品不存在

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/chat/conversations" `
    -Headers $buyerHeaders `
    -ContentType "application/json" `
    -Body '{"productId":999999}'
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
product not found
```

### 5.5 非会话参与人不能查看消息

登录第三个用户并设为已认证：

```powershell
$otherLogin = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/auth/dev-login" `
  -ContentType "application/json" `
  -Body '{"openid":"dev_chat_other_001","role":"USER"}'

$otherToken = $otherLogin.data.token
$otherId = $otherLogin.data.user.id
$otherHeaders = @{ Authorization = "Bearer $otherToken" }
$otherId
```

在 MySQL 中执行：

```sql
UPDATE users
SET auth_status = 'VERIFIED',
    account_status = 'NORMAL',
    nickname = '聊天测试第三人'
WHERE id = 第三人ID;
```

第三人访问会话消息：

```powershell
try {
  Invoke-RestMethod `
    -Method Get `
    -Uri "$baseUrl/chat/conversations/$conversationId/messages" `
    -Headers $otherHeaders
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期：

```text
403
```

### 5.6 空消息不能发送

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/chat/conversations/$conversationId/messages" `
    -Headers $buyerHeaders `
    -ContentType "application/json" `
    -Body '{"content":"   "}'
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期：

```text
400
```

响应消息：

```text
message content is required
```

### 5.7 不存在的会话

```powershell
try {
  Invoke-RestMethod `
    -Method Get `
    -Uri "$baseUrl/chat/conversations/999999/messages" `
    -Headers $buyerHeaders
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
conversation not found
```

## 六、数据库检查

### 6.1 检查会话表

```sql
SELECT id, product_id, buyer_id, seller_id, last_message_content,
       buyer_unread_count, seller_unread_count, status
FROM chat_conversations
WHERE id = 会话ID;
```

预期：

- `product_id` 正确。
- `buyer_id` 正确。
- `seller_id` 正确。
- `last_message_content` 为最后一次发送的内容。
- 未读数字段随已读操作变化。
- `status = ACTIVE`。

### 6.2 检查消息表

```sql
SELECT id, conversation_id, sender_id, receiver_id, content, read_status, create_time
FROM chat_messages
WHERE conversation_id = 会话ID
ORDER BY create_time ASC, id ASC;
```

预期：

- 买家和卖家的消息都存在。
- `sender_id`、`receiver_id` 正确。
- 已读后对应消息 `read_status = READ`。

## 七、Apifox 测试建议

建议新建分组：

```text
王双媛-聊天模块
```

环境变量：

```text
baseUrl = http://localhost:8080
buyerToken = 买家 dev-login 返回的 token
sellerToken = 卖家 dev-login 返回的 token
productId = 测试商品 ID
conversationId = 创建会话返回的 id
```

公共请求头：

```http
Authorization: Bearer {{buyerToken}}
Content-Type: application/json
```

接口列表：

```text
POST {{baseUrl}}/chat/conversations
GET  {{baseUrl}}/chat/conversations?page=1&pageSize=20
GET  {{baseUrl}}/chat/conversations/{{conversationId}}/messages?page=1&pageSize=30
POST {{baseUrl}}/chat/conversations/{{conversationId}}/messages
PUT  {{baseUrl}}/chat/conversations/{{conversationId}}/read
```

## 八、测试结论模板

```text
测试模块：聊天模块

测试接口：
- POST /chat/conversations
- GET /chat/conversations
- GET /chat/conversations/:id/messages
- POST /chat/conversations/:id/messages
- PUT /chat/conversations/:id/read

测试结果：
- 登录鉴权正常
- 学生认证拦截正常
- 创建或复用会话正常
- 商品详情页联系商家流程正常
- 买家发送消息正常
- 卖家回复消息正常
- 会话列表展示最后消息和未读数正常
- 消息列表查询正常
- 标记已读正常
- 非会话参与人访问被拦截
- 空消息和非法参数处理正常

结论：聊天模块接口测试通过
```
