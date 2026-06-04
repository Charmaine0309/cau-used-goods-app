# 王双媛举报申诉处理中状态接口测试说明

## 一、测试范围

本说明用于测试本次新增的管理员受理接口：

```http
POST /admin/reports/:id/processing
POST /admin/appeals/:id/processing
```

## 二、前置条件

服务地址：

```powershell
$baseUrl = "http://localhost:8080"
```

准备管理员 token：

```powershell
$adminLogin = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/auth/dev-login" `
  -ContentType "application/json" `
  -Body '{"openid":"dev_processing_admin_001","role":"ADMIN"}'

$adminToken = $adminLogin.data.token
$adminHeaders = @{ Authorization = "Bearer $adminToken" }
```

准备普通用户 token：

```powershell
$userLogin = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/auth/dev-login" `
  -ContentType "application/json" `
  -Body '{"openid":"dev_processing_user_001","role":"USER"}'

$userToken = $userLogin.data.token
$userHeaders = @{ Authorization = "Bearer $userToken" }
```

将测试账号设置为可用：

```sql
UPDATE users
SET auth_status = 'VERIFIED',
    account_status = 'NORMAL',
    nickname = '处理中测试用户'
WHERE openid = 'dev_processing_user_001';

UPDATE users
SET auth_status = 'VERIFIED',
    account_status = 'NORMAL',
    nickname = '处理中测试管理员',
    role = 'ADMIN'
WHERE openid = 'dev_processing_admin_001';
```

## 三、测试举报进入处理中

先创建一条举报，记录返回的 `reportId`。

```powershell
$report = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/reports" `
  -Headers $userHeaders `
  -ContentType "application/json" `
  -Body '{"targetType":"PRODUCT","targetId":1,"reasonType":"违规商品","description":"处理中状态测试举报"}'

$reportId = $report.data.id
```

管理员标记为处理中：

```powershell
$processingReport = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/admin/reports/$reportId/processing" `
  -Headers $adminHeaders

$processingReport.data.status
```

预期：

```text
PROCESSING
```

再次提交同一目标举报：

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/reports" `
    -Headers $userHeaders `
    -ContentType "application/json" `
    -Body '{"targetType":"PRODUCT","targetId":1,"reasonType":"违规商品","description":"重复举报测试"}'
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期：

```text
409
```

## 四、测试申诉进入处理中

先创建一条申诉，记录返回的 `appealId`。

```powershell
$appeal = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/appeals" `
  -Headers $userHeaders `
  -ContentType "application/json" `
  -Body '{"targetType":"USER","targetId":'$($userLogin.data.user.id)',"reason":"处理中状态测试申诉"}'

$appealId = $appeal.data.id
```

管理员标记为处理中：

```powershell
$processingAppeal = Invoke-RestMethod `
  -Method Post `
  -Uri "$baseUrl/admin/appeals/$appealId/processing" `
  -Headers $adminHeaders

$processingAppeal.data.status
```

预期：

```text
PROCESSING
```

再次提交同一目标申诉：

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/appeals" `
    -Headers $userHeaders `
    -ContentType "application/json" `
    -Body '{"targetType":"USER","targetId":'$($userLogin.data.user.id)',"reason":"重复申诉测试"}'
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期：

```text
409
```

## 五、权限测试

普通用户不能调用管理员受理接口：

```powershell
try {
  Invoke-RestMethod `
    -Method Post `
    -Uri "$baseUrl/admin/appeals/$appealId/processing" `
    -Headers $userHeaders
} catch {
  $_.Exception.Response.StatusCode.value__
}
```

预期：

```text
403
```

## 六、测试结论模板

```text
测试模块：举报申诉处理中状态

测试接口：
- POST /admin/reports/:id/processing
- POST /admin/appeals/:id/processing

测试结果：
- 管理员可以将 PENDING 举报标记为 PROCESSING
- 管理员可以将 PENDING 申诉标记为 PROCESSING
- PROCESSING 后用户不能重复提交同一目标举报
- PROCESSING 后用户不能重复提交同一目标申诉
- 普通用户不能调用管理员受理接口

结论：举报申诉处理中状态接口测试通过
```
