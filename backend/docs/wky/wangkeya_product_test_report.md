# 王珂雅商品相关模块测试说明文档（最终版）

## 1. 测试环境

- 项目名称：CAU 校园二手交易平台
- 测试模块：商品相关模块、图片上传、AI 商品优化、商品统计、收藏/浏览统计联动
- 后端技术：Go + Gin
- 数据库：MySQL，数据库名 `cau_used_goods`
- 本地后端地址：`http://127.0.0.1:8080`
- 测试日期：2026-06-09 至 2026-06-10
- 测试分支：`feature/backend-wangkeya-product-pr`
- 合并基础：已拉取并合并最新 `dev` 分支代码

---

## 2. 测试目的

本次测试用于验证商品相关后端接口在合并最新 `dev` 代码后是否仍能正常运行，并重点检查完整业务流程、接口规范、权限规则、统计字段和跨模块联动是否符合预期。

重点验证内容包括：

1. 商品分类查询；
2. 商品列表展示；
3. 商品关键词搜索；
4. 商品分类筛选；
5. 商品价格区间筛选；
6. 商品成色筛选；
7. 商品排序；
8. 商品详情查询；
9. 商品发布；
10. 我的商品查询；
11. 商品编辑；
12. 商品上下架；
13. 图片上传与商品图片绑定；
14. AI 商品标题/描述优化入口；
15. 商品统计接口及管理员权限控制；
16. 商品删除及删除后的展示逻辑；
17. 浏览量 `viewCount` 是否随商品详情访问增加；
18. 收藏量 `favoriteCount` 是否随收藏关系变化同步；
19. 管理员统计接口中的 `totalViews`、`totalFavorites` 是否与商品浏览/收藏流程一致。

---

## 3. 最新代码同步说明

按照组长要求，先拉取最新后端代码，再合并到本人商品模块分支中重新测试。

执行命令：

```bash
git checkout dev
git pull origin dev
git checkout feature/backend-wangkeya-product-pr
git merge dev
```

合并结果：

```text
Fast-forward
无冲突
```

说明最新 `dev` 代码已成功合并到个人商品模块分支。

---

## 4. 编译测试

### 4.1 执行命令

```bash
go test ./...
```

### 4.2 测试结果

```text
?       cau-used-goods-app/backend/cmd/server              [no test files]
?       cau-used-goods-app/backend/internal/admin          [no test files]
?       cau-used-goods-app/backend/internal/ai             [no test files]
?       cau-used-goods-app/backend/internal/appeal         [no test files]
?       cau-used-goods-app/backend/internal/auth           [no test files]
?       cau-used-goods-app/backend/internal/chat           [no test files]
?       cau-used-goods-app/backend/internal/config         [no test files]
?       cau-used-goods-app/backend/internal/db             [no test files]
?       cau-used-goods-app/backend/internal/favorite       [no test files]
?       cau-used-goods-app/backend/internal/message        [no test files]
?       cau-used-goods-app/backend/internal/middleware     [no test files]
?       cau-used-goods-app/backend/internal/order          [no test files]
?       cau-used-goods-app/backend/internal/product        [no test files]
?       cau-used-goods-app/backend/internal/report         [no test files]
?       cau-used-goods-app/backend/internal/review         [no test files]
?       cau-used-goods-app/backend/internal/sensitive      [no test files]
?       cau-used-goods-app/backend/internal/stats          [no test files]
?       cau-used-goods-app/backend/internal/upload         [no test files]
?       cau-used-goods-app/backend/internal/user           [no test files]
ok      cau-used-goods-app/backend/pkg/jwt                 1.535s
?       cau-used-goods-app/backend/pkg/response            [no test files]
```

### 4.3 结论

后端项目整体编译通过，`pkg/jwt` 测试通过，其余模块暂无单元测试文件。

---

## 5. 数据库结构同步问题记录

### 5.1 问题现象

合并最新 `dev` 后，首次调用开发环境登录接口返回：

```json
{
  "code": 400,
  "message": "find user by openid: Error 1054 (42S22): Unknown column 'token_version' in 'field list'",
  "data": null
}
```

### 5.2 原因分析

最新后端代码中 `users` 表新增了 `token_version` 字段，但本地数据库表结构仍为旧版本，导致查询用户时访问不存在字段。

### 5.3 处理方式

在本地 MySQL 中补充字段：

```sql
USE cau_used_goods;

ALTER TABLE users
ADD COLUMN token_version INT NOT NULL DEFAULT 0 COMMENT 'Token版本号，用于账号状态变更后使旧Token失效';
```

### 5.4 处理结果

补充字段后重新登录成功，返回 `code=0` 和 JWT Token。

---

## 6. 测试账号说明

### 6.1 卖家普通用户账号

```text
openid: wky-product-test-user
role: USER
userId: 6
authStatus: VERIFIED
accountStatus: NORMAL
```

说明：最新代码要求通过学生认证的用户才允许发布商品。因此测试中将该用户设置为 `VERIFIED`。

### 6.2 买家普通用户账号

```text
openid: wky-buyer-test-user
role: USER
userId: 8
authStatus: VERIFIED
accountStatus: NORMAL
```

说明：用于测试浏览商品、收藏商品、收藏状态检查等流程。

### 6.3 管理员账号

```text
openid: wky-admin-test-user
role: ADMIN
userId: 7
accountStatus: NORMAL
```

说明：商品统计接口需要管理员权限，普通用户访问会返回 403，因此使用管理员账号测试统计接口。

---

## 7. conditionLevel 成色筛选逻辑确认

执行命令：

```bash
findstr /N /I "ConditionLevel condition_level" internal\product\handler.go internal\product\service.go internal\product\repository.go
```

关键结果：

```text
internal\product\handler.go:218: conditionLevel := c.Query("conditionLevel")
internal\product\repository.go:259: if input.ConditionLevel != "" {
internal\product\repository.go:260:     where += " AND condition_level = ? "
internal\product\repository.go:261:     args = append(args, input.ConditionLevel)
```

说明：

1. 前端可通过 `conditionLevel` 查询参数传递成色筛选条件；
2. 后端会将该参数传递至 repository 层；
3. SQL 查询中会追加 `AND condition_level = ?`；
4. 当前成色筛选采用精确匹配方式。

---

## 8. 接口测试结果总表

| 编号 | 测试功能 | 请求方式 | 接口路径 | 预期结果 | 实际结果 |
|---|---|---|---|---|---|
| TC-01 | 获取商品分类 | GET | `/categories` | 返回启用分类列表 | 通过 |
| TC-02 | 获取商品列表 | GET | `/products` | 返回分页商品列表 | 通过 |
| TC-03 | 关键词搜索 | GET | `/products?keyword=台灯` | 返回包含关键词的商品 | 通过 |
| TC-04 | 分类筛选 | GET | `/products?categoryId=2` | 返回分类 ID 为 2 的商品 | 通过 |
| TC-05 | 价格区间筛选 | GET | `/products?minPrice=100&maxPrice=200` | 返回价格区间内商品 | 通过 |
| TC-06 | 成色筛选命中 | GET | `/products?...conditionLevel=九成新` | 返回九成新商品 | 通过 |
| TC-07 | 成色筛选不命中 | GET | `/products?...conditionLevel=全新` | 返回 `list:[]`，`total:0` | 通过 |
| TC-08 | 发布商品 | POST | `/products` | 学生认证用户可发布商品 | 通过 |
| TC-09 | 未认证用户发布商品 | POST | `/products` | 返回 403 | 通过 |
| TC-10 | 商品详情 | GET | `/products/:id` | 返回商品详情 | 通过 |
| TC-11 | 我的商品 | GET | `/products/my` | 返回当前用户发布的商品 | 通过 |
| TC-12 | 编辑商品 | PUT | `/products/:id` | 商品信息修改成功 | 通过 |
| TC-13 | 下架商品 | PUT | `/products/:id/status` | 状态改为 `OFF_SHELF` | 通过 |
| TC-14 | 下架后详情查询 | GET | `/products/:id` | 普通详情接口不展示下架商品 | 通过 |
| TC-15 | 重新上架 | PUT | `/products/:id/status` | 状态改为 `ON_SALE` | 通过 |
| TC-16 | 上传图片 | POST | `/upload/image` | 返回图片 URL | 通过 |
| TC-17 | 绑定商品图片 | POST | `/products/:id/images` | 图片绑定成功 | 通过 |
| TC-18 | AI 优化入口 | POST | `/ai/optimize-product` | 未配置 Key 时返回可控错误 | 通过 |
| TC-19 | 普通用户访问统计接口 | GET | `/stats/products/*` | 返回 403，需要管理员权限 | 通过 |
| TC-20 | 管理员访问统计接口 | GET | `/stats/products/*` | 返回统计数据 | 通过 |
| TC-21 | 删除商品 | DELETE | `/products/:id` | 删除成功 | 通过 |
| TC-22 | 删除后详情查询 | GET | `/products/:id` | 返回 `product not found` | 通过 |
| TC-23 | 删除后列表搜索 | GET | `/products?keyword=宿舍自用台灯测试版` | 返回空列表 | 通过 |
| TC-24 | 买家浏览商品详情 | GET | `/products/7` | 期望 `viewCount` 增加 | 未通过，仍为 0 |
| TC-25 | 买家收藏商品 | POST | `/favorites` | 返回 `favorited=true` | 通过 |
| TC-26 | 检查收藏状态 | GET | `/favorites/check?productId=7` | 返回 `favorited=true` | 通过 |
| TC-27 | 收藏后商品收藏数 | GET | `/products/7` | 期望 `favoriteCount` 增加 | 未通过，仍为 0 |
| TC-28 | 收藏后管理员统计 | GET | `/stats/products/overview`、`/stats/products/category-distribution` | 期望 `totalFavorites` 增加 | 未通过，仍为 0 |

---

## 9. 核心功能详细测试记录

### 9.1 获取商品分类

请求：

```bash
curl "http://127.0.0.1:8080/categories"
```

结果：返回 6 个启用分类，包括教材资料、电子产品、生活用品、服饰鞋包、运动户外、其他。

结论：通过。

### 9.2 获取商品列表

请求：

```bash
curl "http://127.0.0.1:8080/products"
```

结果：返回 `code=0`，`data.list` 为数组，包含商品列表、页码、分页大小和总数。

结论：通过。

### 9.3 关键词搜索

请求：

```bash
curl "http://127.0.0.1:8080/products?keyword=台灯"
```

结果：返回包含“台灯”的商品。

结论：通过。

### 9.4 分类筛选

请求：

```bash
curl "http://127.0.0.1:8080/products?categoryId=2"
```

结果：返回电子产品分类下的商品。

结论：通过。

### 9.5 价格区间筛选

请求：

```bash
curl "http://127.0.0.1:8080/products?minPrice=100&maxPrice=200"
```

结果：返回价格在 100 到 200 之间的商品。

结论：通过。

### 9.6 成色筛选命中

请求：

```bash
curl "http://127.0.0.1:8080/products?keyword=台灯&categoryId=2&minPrice=100&maxPrice=200&conditionLevel=九成新&sort=price_asc"
```

结果：返回 `total=1`，商品成色为 `九成新`。

结论：通过，说明 `conditionLevel=九成新` 生效。

### 9.7 成色筛选不命中

请求：

```bash
curl "http://127.0.0.1:8080/products?keyword=台灯&categoryId=2&minPrice=100&maxPrice=200&conditionLevel=全新&sort=price_asc"
```

结果：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [],
    "page": 1,
    "pageSize": 10,
    "total": 0
  }
}
```

结论：通过，说明成色筛选真正参与查询，并且空列表返回 `[]` 而不是 `null`。

### 9.8 发布商品

请求：

```bash
curl -X POST http://127.0.0.1:8080/products ^
-H "Content-Type: application/json" ^
-H "Authorization: Bearer <token>" ^
-d "{"categoryId":2,"title":"宿舍自用台灯测试版","description":"九成新台灯，适合宿舍学习使用","originalPrice":199,"price":150,"conditionLevel":"九成新","meetLocation":"图书馆门口"}"
```

结果：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 6
  },
  "timestamp": "2026-06-09 18:36:34"
}
```

结论：通过。

### 9.9 未认证用户发布商品

请求：使用 `authStatus=UNVERIFIED` 的普通用户发布商品。

结果：

```json
{
  "code": 403,
  "message": "student verification required",
  "data": null
}
```

结论：通过，说明学生认证限制生效。

### 9.10 商品详情

请求：

```bash
curl "http://127.0.0.1:8080/products/6"
```

结果：返回商品标题、描述、价格、成色、交易地点、状态和图片列表。

结论：通过。

### 9.11 我的商品

请求：

```bash
curl "http://127.0.0.1:8080/products/my" ^
-H "Authorization: Bearer <token>"
```

结果：返回当前用户发布的商品。

结论：通过。

### 9.12 编辑商品

请求：

```bash
curl -X PUT http://127.0.0.1:8080/products/6 ^
-H "Content-Type: application/json" ^
-H "Authorization: Bearer <token>" ^
-d "{"categoryId":2,"title":"宿舍自用台灯测试版-已修改","description":"修改后的九成新台灯，功能正常","originalPrice":199,"price":140,"conditionLevel":"九成新","meetLocation":"东区食堂门口"}"
```

结果：返回 `code=0`，再次查询详情后，商品标题、价格、交易地点均已更新。

结论：通过。

### 9.13 商品下架与重新上架

下架请求：

```bash
curl -X PUT http://127.0.0.1:8080/products/6/status ^
-H "Content-Type: application/json" ^
-H "Authorization: Bearer <token>" ^
-d "{"status":"OFF_SHELF"}"
```

结果：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 6,
    "status": "OFF_SHELF"
  }
}
```

下架后普通详情查询返回：

```json
{
  "code": 404,
  "message": "product not found",
  "data": null
}
```

重新上架后再次查询详情，商品状态恢复为 `ON_SALE`。

结论：通过。

### 9.14 图片上传

请求：

```bash
curl -X POST http://127.0.0.1:8080/upload/image ^
-H "Authorization: Bearer <token>" ^
-F "file=@C:\Users\18472\Desktop\test.png"
```

结果：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "url": "/uploads/products/1781001668926493800.png"
  },
  "timestamp": "2026-06-09 18:41:08"
}
```

结论：通过。

### 9.15 商品图片绑定

请求：

```bash
curl -X POST http://127.0.0.1:8080/products/6/images ^
-H "Content-Type: application/json" ^
-H "Authorization: Bearer <token>" ^
-d "{"images":["/uploads/products/1781001668926493800.png"]}"
```

结果：返回 `code=0`，再次查询商品详情，`images` 中已包含该图片 URL。

结论：通过。

### 9.16 AI 商品优化入口

请求：

```bash
curl -X POST http://127.0.0.1:8080/ai/optimize-product ^
-H "Content-Type: application/json" ^
-H "Authorization: Bearer <token>" ^
-d "{"title":"高数教材","description":"正版教材，九成新，适合期末复习"}"
```

结果：

```json
{
  "code": 500,
  "message": "AI服务暂不可用：未配置 API Key",
  "data": null,
  "timestamp": "2026-06-09 18:38:51"
}
```

结论：通过。当前本地测试环境未配置 AI API Key，接口返回可控错误提示，后端服务未崩溃。

### 9.17 普通用户访问统计接口

普通用户访问以下接口：

```text
GET /stats/products/overview
GET /stats/products/category-distribution
GET /stats/products/status-distribution
GET /stats/products/trend?days=7
```

结果均为：

```json
{
  "code": 403,
  "message": "需要管理员权限",
  "data": null
}
```

结论：通过，统计接口权限控制正常。

### 9.18 管理员访问统计接口

管理员访问统计接口均返回 `code=0`。

商品总览统计示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "totalProducts": 7,
    "onSaleProducts": 4,
    "offShelfProducts": 0,
    "lockedProducts": 0,
    "soldProducts": 1,
    "deletedProducts": 2,
    "totalViews": 0,
    "totalFavorites": 0,
    "averagePrice": 66.4
  },
  "timestamp": "2026-06-10 10:12:32"
}
```

分类分布统计示例：

```json
{
  "categoryId": 2,
  "categoryName": "电子产品",
  "productCount": 4,
  "onSaleCount": 2,
  "averagePrice": 135,
  "totalViews": 0,
  "totalFavorites": 0
}
```

结论：接口访问通过，但浏览量和收藏量统计存在联动问题，详见第 10 节。

### 9.19 删除商品

请求：

```bash
curl -X DELETE http://127.0.0.1:8080/products/6 ^
-H "Authorization: Bearer <token>"
```

结果：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 6
  },
  "timestamp": "2026-06-09 18:47:01"
}
```

删除后查询详情：

```json
{
  "code": 404,
  "message": "product not found",
  "data": null
}
```

删除后搜索商品：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [],
    "page": 1,
    "pageSize": 10,
    "total": 0
  }
}
```

结论：通过。

---

## 10. 浏览量与收藏量联动专项测试

### 10.1 测试商品

重新发布测试商品：

```text
商品ID：7
标题：浏览收藏统计测试商品
分类：电子产品
价格：120
成色：九成新
卖家用户ID：6
```

### 10.2 浏览量测试

使用买家 token 访问商品详情：

```bash
curl "http://127.0.0.1:8080/products/7" ^
-H "Authorization: Bearer <buyer_token>"
```

返回商品详情中：

```json
"viewCount": 0
```

再次不带 token 查询商品详情：

```bash
curl "http://127.0.0.1:8080/products/7"
```

返回商品详情中：

```json
"viewCount": 0
```

结论：未通过。商品详情接口访问后，`viewCount` 未增加。

### 10.3 收藏接口测试

买家收藏商品：

```bash
curl -X POST http://127.0.0.1:8080/favorites ^
-H "Content-Type: application/json" ^
-H "Authorization: Bearer <buyer_token>" ^
-d "{"productId":7}"
```

返回：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "favorited": true
  },
  "timestamp": "2026-06-10 10:12:08"
}
```

检查收藏状态：

```bash
curl "http://127.0.0.1:8080/favorites/check?productId=7" ^
-H "Authorization: Bearer <buyer_token>"
```

返回：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "favorited": true
  },
  "timestamp": "2026-06-10 10:12:14"
}
```

结论：收藏接口本身通过，收藏关系创建成功。

### 10.4 数据库验证

查询收藏表：

```sql
SELECT * FROM favorites WHERE product_id = 7;
```

结果：

```text
+----+---------+------------+---------------------+------------+
| id | user_id | product_id | create_time         | is_deleted |
+----+---------+------------+---------------------+------------+
|  1 |       8 |          7 | 2026-06-10 10:12:08 |          0 |
+----+---------+------------+---------------------+------------+
```

说明买家 `user_id=8` 已收藏商品 `product_id=7`。

查询商品表：

```sql
SELECT id, title, view_count, favorite_count
FROM products
WHERE id = 7;
```

结果：

```text
+----+--------------------------------+------------+----------------+
| id | title                          | view_count | favorite_count |
+----+--------------------------------+------------+----------------+
|  7 | 浏览收藏统计测试商品           |          0 |              0 |
+----+--------------------------------+------------+----------------+
```

说明收藏关系存在，但商品表冗余统计字段未同步更新。

### 10.5 管理员统计验证

管理员访问商品总览统计：

```bash
curl "http://127.0.0.1:8080/stats/products/overview" ^
-H "Authorization: Bearer <admin_token>"
```

返回中：

```json
"totalViews": 0,
"totalFavorites": 0
```

管理员访问分类分布统计：

```bash
curl "http://127.0.0.1:8080/stats/products/category-distribution" ^
-H "Authorization: Bearer <admin_token>"
```

电子产品分类返回中：

```json
"totalViews": 0,
"totalFavorites": 0
```

结论：管理员统计接口可以正常访问，但其中浏览量与收藏量字段未反映实际浏览/收藏行为。

---

## 11. 已发现问题与处理记录

### 11.1 本地数据库缺少 `token_version` 字段

- 问题：合并最新代码后，登录接口报 `Unknown column 'token_version' in 'field list'`。
- 原因：本地数据库表结构未同步最新 `dev`。
- 处理：在 `users` 表中补充 `token_version` 字段。
- 结果：登录恢复正常。

### 11.2 未认证用户无法发布或收藏商品

- 问题：普通用户发布商品或收藏商品返回 `403 student verification required`。
- 原因：最新代码要求相关业务操作用户完成学生认证。
- 处理：将测试用户 `auth_status` 设置为 `VERIFIED` 后重新测试。
- 结果：认证用户可正常发布商品和收藏商品。

### 11.3 下架商品普通详情接口不可见

- 现象：商品状态改为 `OFF_SHELF` 后，普通详情接口返回 `product not found`。
- 判断：该行为符合当前业务逻辑，下架商品不对普通用户展示。
- 处理：记录在测试说明中，无需修改代码。

### 11.4 AI 未配置 API Key

- 现象：调用 AI 优化接口返回 `AI服务暂不可用：未配置 API Key`。
- 判断：本地测试环境未配置真实 AI API Key，属于可控错误。
- 处理：记录在测试说明中，不影响用户手动发布商品。

### 11.5 浏览量统计未联动更新

- 现象：买家访问 `GET /products/7` 商品详情后，商品详情中的 `viewCount` 仍为 0。
- 判断：商品详情接口当前未触发 `products.view_count` 自增，或浏览量统计逻辑尚未接入详情查询流程。
- 处理建议：在明确责任范围后，可考虑在商品详情查询流程中增加浏览量更新逻辑，并注意避免卖家本人浏览、重复刷新等场景造成异常统计。

### 11.6 收藏量统计未联动更新

- 现象：买家调用 `POST /favorites` 收藏商品成功，`GET /favorites/check?productId=7` 返回 `favorited=true`，数据库 `favorites` 表也存在有效收藏记录，但商品详情中的 `favoriteCount` 仍为 0，管理员统计接口中的 `totalFavorites` 仍为 0。
- 判断：收藏关系创建成功，但 `products.favorite_count` 未随收藏/取消收藏流程同步更新；统计接口读取的收藏量也未体现实际收藏关系。
- 处理建议：在明确责任范围后，可考虑将收藏表操作与商品 `favorite_count` 更新放入同一事务中，保证收藏关系与商品冗余统计字段一致。

---

## 12. 测试结论

在合并最新 `dev` 分支代码后，商品相关模块编译通过，分类、商品列表、搜索筛选、成色筛选、商品详情、发布商品、我的商品、编辑商品、上下架、图片上传、图片绑定、AI 优化入口、商品统计权限、管理员统计接口和商品删除等核心功能均已完成验证。

本次专项补充测试发现：浏览量与收藏量统计存在联动一致性问题。具体表现为：商品详情访问后 `viewCount` 未增加；收藏关系创建成功后 `favoriteCount` 未增加；管理员统计接口中的 `totalViews`、`totalFavorites` 也仍为 0。该问题已在测试说明中记录，后续如需修改代码，应结合商品详情、收藏模块和统计模块统一处理，并注意事务一致性。
