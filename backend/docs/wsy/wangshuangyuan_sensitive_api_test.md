# 王双媛敏感词管理接口测试说明

## 一、测试范围

本文档用于记录管理员敏感词管理模块的接口测试过程与结果。

测试接口：

```http
GET /admin/sensitive-words
POST /admin/sensitive-words
PUT /admin/sensitive-words/:id
DELETE /admin/sensitive-words/:id
```

测试重点：

- 管理员可以新增、查询、更新、禁用敏感词。
- 普通用户不能访问敏感词管理接口。
- 非法参数可以被正确拦截。
- 不存在的敏感词 ID 可以返回正确错误。
- 新增、更新、禁用敏感词时会写入 `admin_logs`。

## 二、测试前置条件

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

本模块依赖数据表：

```text
users
sensitive_words
admin_logs
```

敏感词字段约定：

```text
wordType: FORBIDDEN / RISK
status: ENABLED / DISABLED
```

说明：

- `FORBIDDEN` 表示禁止类敏感词。
- `RISK` 表示风险类敏感词。
- `DELETE /admin/sensitive-words/:id` 当前按禁用处理，即将 `status` 设置为 `DISABLED`，不是物理删除。

## 三、PowerShell 测试命令

### 3.1 设置基础地址

```powershell
$baseUrl = "http://localhost:8080"
```

### 3.2 获取管理员 Token

```powershell
$adminLogin = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/auth/dev-login" `
  -ContentType "application/json" `
  -Body '{"openid":"dev_admin_word_001","role":"ADMIN"}'

$adminToken = $adminLogin.data.token
$adminHeaders = @{ Authorization = "Bearer $adminToken" }

$adminLogin
```

预期结果：

- 返回 `code = 0`。
- `data.token` 有值。
- 当前用户角色为 `ADMIN`。

### 3.3 新增敏感词

```powershell
$createResult = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/admin/sensitive-words" `
  -Headers $adminHeaders `
  -ContentType "application/json" `
  -Body '{"word":"测试敏感词001","wordType":"FORBIDDEN","status":"ENABLED"}'

$wordId = $createResult.data.id
$createResult
$wordId
```

预期结果：

- 返回 `code = 0`。
- 返回新建敏感词 ID。
- `admin_logs` 中新增 `CREATE_WORD` 操作日志。

### 3.4 查询敏感词列表

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "$baseUrl/admin/sensitive-words?page=1&pageSize=20" `
  -Headers $adminHeaders
```

预期结果：

- 返回 `code = 0`。
- 返回分页数据。
- `items` 中包含刚创建的敏感词。

### 3.5 按状态筛选

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "$baseUrl/admin/sensitive-words?status=ENABLED&page=1&pageSize=20" `
  -Headers $adminHeaders
```

预期结果：

- 返回 `code = 0`。
- 返回结果中敏感词状态为 `ENABLED`。

### 3.6 按类型筛选

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "$baseUrl/admin/sensitive-words?wordType=FORBIDDEN&page=1&pageSize=20" `
  -Headers $adminHeaders
```

预期结果：

- 返回 `code = 0`。
- 返回结果中敏感词类型为 `FORBIDDEN`。

### 3.7 按关键字搜索

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "$baseUrl/admin/sensitive-words?keyword=测试&page=1&pageSize=20" `
  -Headers $adminHeaders
```

预期结果：

- 返回 `code = 0`。
- 返回 `word` 包含 `测试` 的敏感词记录。

### 3.8 更新敏感词

```powershell
$updateResult = Invoke-RestMethod `
  -Method Put `
  -Uri "$baseUrl/admin/sensitive-words/$wordId" `
  -Headers $adminHeaders `
  -ContentType "application/json" `
  -Body '{"word":"测试敏感词001-已更新","wordType":"RISK","status":"ENABLED"}'

$updateResult
```

预期结果：

- 返回 `code = 0`。
- 返回 `updated = true`。
- `admin_logs` 中新增 `UPDATE_WORD` 操作日志。

### 3.9 删除/禁用敏感词

```powershell
$deleteResult = Invoke-RestMethod `
  -Method Delete `
  -Uri "$baseUrl/admin/sensitive-words/$wordId" `
  -Headers $adminHeaders

$deleteResult
```

预期结果：

- 返回 `code = 0`。
- 返回 `deleted = true`。
- 数据库中该敏感词状态变为 `DISABLED`。
- `admin_logs` 中新增 `DELETE_WORD` 操作日志。

### 3.10 查询禁用状态确认

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "$baseUrl/admin/sensitive-words?status=DISABLED&page=1&pageSize=20" `
  -Headers $adminHeaders
```

预期结果：

- 返回 `code = 0`。
- 可以查询到已禁用的敏感词。
- 该敏感词 `status = DISABLED`。

## 四、权限测试

### 4.1 普通用户不能访问敏感词管理接口

```powershell
$userLogin = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/auth/dev-login" `
  -ContentType "application/json" `
  -Body '{"openid":"dev_user_word_001","role":"USER"}'

$userToken = $userLogin.data.token
$userHeaders = @{ Authorization = "Bearer $userToken" }

try {
  Invoke-RestMethod `
    -Method Get `
    -Uri "$baseUrl/admin/sensitive-words" `
    -Headers $userHeaders
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期结果：

```text
403
```

说明：

- 如果返回 `401`，通常表示 token 未正确传入或已失效。
- 普通用户携带有效 token 访问管理员接口时，应返回 `403`。

## 五、异常参数测试

### 5.1 新增敏感词缺少 word

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/admin/sensitive-words" `
    -Headers $adminHeaders `
    -ContentType "application/json" `
    -Body '{"wordType":"FORBIDDEN","status":"ENABLED"}'
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期结果：

```text
400
```

### 5.2 非法敏感词类型

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/admin/sensitive-words" `
    -Headers $adminHeaders `
    -ContentType "application/json" `
    -Body '{"word":"非法类型测试","wordType":"INVALID","status":"ENABLED"}'
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期结果：

```text
400
```

### 5.3 非法敏感词状态

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/admin/sensitive-words" `
    -Headers $adminHeaders `
    -ContentType "application/json" `
    -Body '{"word":"非法状态测试","wordType":"FORBIDDEN","status":"INVALID"}'
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期结果：

```text
400
```

### 5.4 不存在的敏感词 ID

```powershell
try {
  Invoke-RestMethod `
    -Method Delete `
    -Uri "$baseUrl/admin/sensitive-words/999999" `
    -Headers $adminHeaders
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期结果：

```text
404
```

## 六、管理员日志校验

新增、更新、删除/禁用敏感词后，查询 `admin_logs`：

```sql
SELECT id, admin_id, operation_type, target_type, target_id, description, ip_address, create_time
FROM admin_logs
WHERE target_type = 'WORD'
ORDER BY id DESC;
```

预期结果：

```text
operation_type 包含 CREATE_WORD / UPDATE_WORD / DELETE_WORD
target_type = WORD
target_id 为对应敏感词 ID
```

说明：

- `GET /admin/sensitive-words` 查询列表不写入管理员操作日志。
- `POST /admin/sensitive-words` 写入 `CREATE_WORD`。
- `PUT /admin/sensitive-words/:id` 写入 `UPDATE_WORD`。
- `DELETE /admin/sensitive-words/:id` 写入 `DELETE_WORD`。

## 七、测试结论

测试模块：管理员敏感词管理模块

测试接口：

```text
GET /admin/sensitive-words
POST /admin/sensitive-words
PUT /admin/sensitive-words/:id
DELETE /admin/sensitive-words/:id
```

测试结果：

- 管理员登录成功，能够获取有效 token。
- 管理员可以新增敏感词。
- 管理员可以分页查询敏感词列表。
- 管理员可以按状态、类型、关键字筛选敏感词。
- 管理员可以更新敏感词内容、类型和状态。
- 管理员可以删除/禁用敏感词。
- 普通用户访问敏感词管理接口时返回 `403`。
- 缺少必填参数时返回 `400`。
- 非法 `wordType` 或非法 `status` 时返回 `400`。
- 不存在的敏感词 ID 返回 `404`。
- 新增、更新、删除/禁用敏感词时均已写入 `admin_logs`。

结论：

```text
管理员敏感词管理模块接口测试通过。
```
