# 王双媛分类管理与申诉统计接口测试说明

## 一、测试范围

本说明用于测试本次新增或修复内容：

```http
GET /stats/appeals/overview
GET /admin/categories
POST /admin/categories
PUT /admin/categories/:id
PUT /admin/categories/:id/status
DELETE /admin/categories/:id
GET /categories
```

同时确认：

```text
go test ./... 不再因为 review 模块的评分格式化问题失败
```

## 二、测试前置条件

本地服务地址：

```text
http://localhost:8080
```

需要管理员 token。

```powershell
$baseUrl = "http://localhost:8080"

$adminLogin = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/auth/dev-login" `
  -ContentType "application/json" `
  -Body '{"openid":"dev_category_admin_001","role":"ADMIN"}'

$adminToken = $adminLogin.data.token
$adminHeaders = @{ Authorization = "Bearer $adminToken" }
```

如果管理员权限未生效，在 MySQL 中确认：

```sql
UPDATE users
SET role = 'ADMIN',
    account_status = 'NORMAL'
WHERE openid = 'dev_category_admin_001';
```

## 三、申诉统计测试

### 3.1 查询申诉统计

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "$baseUrl/stats/appeals/overview" `
  -Headers $adminHeaders
```

预期字段：

```text
totalAppeals
pendingAppeals
processingAppeals
approvedAppeals
rejectedAppeals
closedAppeals
productAppeals
userAppeals
orderAppeals
reportAppeals
```

### 3.2 普通用户不能访问申诉统计

```powershell
$userLogin = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/auth/dev-login" `
  -ContentType "application/json" `
  -Body '{"openid":"dev_category_user_001","role":"USER"}'

$userHeaders = @{ Authorization = "Bearer $($userLogin.data.token)" }

try {
  Invoke-RestMethod `
    -Method Get `
    -Uri "$baseUrl/stats/appeals/overview" `
    -Headers $userHeaders
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期：

```text
403
```

## 四、分类管理测试

### 4.1 管理员查看全部分类

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "$baseUrl/admin/categories" `
  -Headers $adminHeaders
```

预期：

- 返回全部分类，包括 `ENABLED` 和 `DISABLED`。
- 按 `sortOrder ASC, id ASC` 排序。

### 4.2 管理员按状态筛选分类

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "$baseUrl/admin/categories?status=ENABLED" `
  -Headers $adminHeaders
```

预期：

- 只返回 `status = ENABLED` 的分类。

### 4.3 新增分类

```powershell
$category = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/admin/categories" `
  -Headers $adminHeaders `
  -ContentType "application/json" `
  -Body '{"name":"测试分类","parentId":0,"sortOrder":888,"status":"ENABLED"}'

$categoryId = $category.data.id
$categoryId
```

预期：

- 返回新分类 ID。
- 数据库中新增一条分类。
- `admin_logs` 中新增 `CREATE_CATEGORY` 日志，`target_type = CATEGORY`。

### 4.4 修改分类

```powershell
Invoke-RestMethod `
  -Method Put `
  -Uri "$baseUrl/admin/categories/$categoryId" `
  -Headers $adminHeaders `
  -ContentType "application/json" `
  -Body '{"name":"测试分类-修改","parentId":0,"sortOrder":889,"status":"ENABLED"}'
```

预期：

- 返回分类 ID。
- 分类名称和排序更新。
- `admin_logs` 中新增 `UPDATE_CATEGORY` 日志，`target_type = CATEGORY`。

### 4.5 禁用分类

```powershell
Invoke-RestMethod `
  -Method Put `
  -Uri "$baseUrl/admin/categories/$categoryId/status" `
  -Headers $adminHeaders `
  -ContentType "application/json" `
  -Body '{"status":"DISABLED"}'
```

预期：

```text
data.status = DISABLED
```

并写入 `STATUS_CATEGORY` 日志。

### 4.6 普通分类列表不展示禁用分类

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "$baseUrl/categories"
```

预期：

- 不包含刚被禁用的分类。

### 4.7 启用分类

```powershell
Invoke-RestMethod `
  -Method Put `
  -Uri "$baseUrl/admin/categories/$categoryId/status" `
  -Headers $adminHeaders `
  -ContentType "application/json" `
  -Body '{"status":"ENABLED"}'
```

预期：

```text
data.status = ENABLED
```

并写入 `STATUS_CATEGORY` 日志。

### 4.8 删除分类

```powershell
Invoke-RestMethod `
  -Method Delete `
  -Uri "$baseUrl/admin/categories/$categoryId" `
  -Headers $adminHeaders
```

预期：

```text
data.status = DISABLED
```

说明：

- 删除分类是逻辑删除，即禁用分类。
- 不物理删除，避免影响已有商品外键。
- 删除操作写入 `DELETE_CATEGORY` 日志。

## 五、异常场景测试

### 5.1 普通用户不能管理分类

```powershell
try {
  Invoke-RestMethod `
    -Method Get `
    -Uri "$baseUrl/admin/categories" `
    -Headers $userHeaders
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期：

```text
403
```

### 5.2 分类名称不能为空

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/admin/categories" `
    -Headers $adminHeaders `
    -ContentType "application/json" `
    -Body '{"name":"","parentId":0,"sortOrder":1,"status":"ENABLED"}'
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期：

```text
400
```

### 5.3 非法分类状态

```powershell
try {
  Invoke-RestMethod `
    -Method Put `
    -Uri "$baseUrl/admin/categories/$categoryId/status" `
    -Headers $adminHeaders `
    -ContentType "application/json" `
    -Body '{"status":"INVALID"}'
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期：

```text
400
```

## 六、编译检查

```powershell
cd D:\cau-used-goods-app\backend
go test ./...
```

预期：

```text
无 FAIL
```

如果遇到 Go 构建缓存权限问题，可以临时指定缓存目录：

```powershell
$env:GOCACHE='D:\cau-used-goods-app\.gocache'
go test ./...
```

测试后不要提交 `.gocache`。

## 七、测试结论模板

```text
测试模块：分类管理与申诉统计

测试接口：
- GET /stats/appeals/overview
- GET /admin/categories
- POST /admin/categories
- PUT /admin/categories/:id
- PUT /admin/categories/:id/status
- DELETE /admin/categories/:id
- GET /categories

测试结果：
- 申诉统计查询正常
- 管理员分类列表查询正常
- 新增分类正常
- 修改分类正常
- 启用/禁用分类正常
- 删除分类按禁用处理正常
- 普通分类列表不展示禁用分类
- 普通用户不能访问后台分类管理
- 非法参数处理正常
- go test ./... 编译通过

结论：分类管理与申诉统计接口测试通过
```
