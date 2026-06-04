# 王双媛分类管理与申诉统计补充说明

## 一、改动范围

本次补充包含三部分：

- 后台申诉统计接口
- 管理员商品分类管理接口
- 评价通知编译问题修复

涉及代码：

```text
backend/internal/stats/
backend/internal/product/
backend/internal/review/service.go
backend/cmd/server/main.go
backend/docs/wsy/wangshuangyuan_stats_api_test.md
```

## 二、后台申诉统计

新增接口：

```http
GET /stats/appeals/overview
```

权限：

```text
登录 + role=ADMIN
```

统计来源表：

```text
appeals
```

返回字段：

```text
totalAppeals       申诉总数
pendingAppeals     待处理申诉数
processingAppeals  处理中申诉数
approvedAppeals    已通过申诉数
rejectedAppeals    已驳回申诉数
closedAppeals      已关闭申诉数
productAppeals     商品申诉数
userAppeals        用户账号申诉数
orderAppeals       订单申诉数
reportAppeals      举报处理申诉数
```

用途：

- 后台首页展示申诉处理概况。
- 管理员快速查看待处理申诉数量。
- 统计不同申诉类型占比。

## 三、管理员商品分类管理

原有接口：

```http
GET /categories
```

说明：

- 仅返回 `status = ENABLED` 的分类。
- 用于普通前端展示、发布商品时选择分类。

本次新增管理员接口：

```http
GET    /admin/categories
POST   /admin/categories
PUT    /admin/categories/:id
PUT    /admin/categories/:id/status
DELETE /admin/categories/:id
```

权限：

```text
登录 + role=ADMIN
```

### 3.1 查看全部分类

```http
GET /admin/categories
```

支持按状态筛选：

```http
GET /admin/categories?status=ENABLED
GET /admin/categories?status=DISABLED
```

返回字段：

```text
id
name
parentId
sortOrder
status
```

### 3.2 新增分类

```http
POST /admin/categories
```

请求体：

```json
{
  "name": "数码配件",
  "parentId": 0,
  "sortOrder": 60,
  "status": "ENABLED"
}
```

规则：

- `name` 必填，最多 50 个字符。
- `status` 可选，默认为 `ENABLED`。
- `status` 只能是 `ENABLED` 或 `DISABLED`。

### 3.3 修改分类

```http
PUT /admin/categories/:id
```

请求体：

```json
{
  "name": "数码电子",
  "parentId": 0,
  "sortOrder": 20,
  "status": "ENABLED"
}
```

规则：

- 不能把分类父级设为自己。
- 修改会更新 `update_time`。

### 3.4 启用或禁用分类

```http
PUT /admin/categories/:id/status
```

请求体：

```json
{
  "status": "DISABLED"
}
```

说明：

- `ENABLED` 表示前端可见、可选择。
- `DISABLED` 表示前端普通分类列表不展示。

### 3.5 删除分类

```http
DELETE /admin/categories/:id
```

当前删除不是物理删除，而是：

```text
status = DISABLED
```

原因：

- `products.category_id` 外键引用 `categories.id`。
- 如果物理删除分类，已有商品数据可能受影响。
- 禁用分类更利于数据追溯。

## 四、评价通知编译修复

修复文件：

```text
backend/internal/review/service.go
```

问题：

```text
fmt.Sprintf 使用 %f 格式化 int 类型的 input.Rating
```

修复：

```go
fmt.Sprintf("您的订单收到%d星评价", input.Rating)
```

影响：

- 修复 `go test ./...` 编译失败。
- 不改变评价业务逻辑。

## 五、前端联调建议

后台分类管理页可以对应：

```text
分类列表
新增分类
编辑分类
启用/禁用分类
删除分类
```

普通商品发布页继续调用：

```http
GET /categories
```

后台统计页新增一个申诉统计卡片：

```http
GET /stats/appeals/overview
```

建议展示：

```text
待处理申诉
已通过申诉
已驳回申诉
商品申诉
账号申诉
订单申诉
举报申诉
```
