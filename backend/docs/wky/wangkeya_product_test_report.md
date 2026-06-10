# 王珂雅商品相关模块测试说明文档（修复版）

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

在组内退回意见中，明确指出收藏数和浏览数不能增加，需要修复该部分代码。因此本次补充测试重点验证：

1. 商品详情访问后 `products.view_count` 是否自动增加；
2. 收藏商品后 `products.favorite_count` 是否自动增加；
3. 取消收藏后 `products.favorite_count` 是否自动减少且不小于 0；
4. `favorites` 表与 `products.favorite_count` 是否保持一致；
5. 管理员统计接口 `totalViews`、`totalFavorites` 是否同步变化。

---

## 3. 最新代码同步与修改说明

本次在原商品文档提交基础上，补充修改了浏览量和收藏量统计逻辑。

### 3.1 修改文件

```text
internal/product/repository.go
internal/product/service.go
internal/favorite/repository.go
docs/wky/wangkeya_product_api_doc.md
docs/wky/wangkeya_product_test_report.md
```

### 3.2 代码修改点

#### 3.2.1 浏览量修复

在商品仓库层新增 `IncrementViewCount` 方法，访问商品详情时对在售商品执行：

```sql
UPDATE products
SET view_count = view_count + 1,
    update_time = CURRENT_TIMESTAMP
WHERE id = ?
  AND is_deleted = 0
  AND status = 'ON_SALE'
```

在商品服务层 `GetProductByID` 中，先调用 `IncrementViewCount`，再查询商品详情，保证详情接口返回时浏览量已经更新。

#### 3.2.2 收藏量修复

修改收藏仓库层 `Add` 和 `Remove` 方法，将收藏关系表与商品收藏数字段放入同一事务中处理：

- 新增收藏或恢复收藏时，`products.favorite_count + 1`；
- 取消收藏时，`products.favorite_count - 1`；
- 通过 `CASE WHEN favorite_count > 0 THEN favorite_count - 1 ELSE 0 END` 保证收藏数不会小于 0；
- 使用事务和 `FOR UPDATE` 避免收藏关系与商品统计字段不一致。

---

## 4. 编译测试

### 4.1 执行命令

```bash
gofmt -w internal\product\repository.go internal\product\service.go internal\favorite\repository.go
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
ok      cau-used-goods-app/backend/pkg/jwt                 (cached)
?       cau-used-goods-app/backend/pkg/response            [no test files]
```

### 4.3 结论

代码格式化成功，`go test ./...` 通过，浏览量与收藏量相关代码无编译错误。

---

## 5. 测试账号说明

### 5.1 卖家普通用户账号

```text
openid: wky-product-test-user
role: USER
userId: 6
authStatus: VERIFIED
accountStatus: NORMAL
```

### 5.2 买家普通用户账号

```text
openid: wky-buyer-test-user
role: USER
userId: 8
authStatus: VERIFIED
accountStatus: NORMAL
```

### 5.3 管理员账号

```text
openid: wky-admin-test-user
role: ADMIN
userId: 7
accountStatus: NORMAL
```

---

## 6. 接口测试结果总表

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
| TC-10 | 商品详情 | GET | `/products/:id` | 返回商品详情且 viewCount 增加 | 通过 |
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
| TC-24 | 买家浏览商品详情 | GET | `/products/7` | `viewCount` 增加 | 通过，`view_count` 由 0 增加为 1 |
| TC-25 | 买家取消收藏商品 | DELETE | `/favorites/7` | 返回 `favorited=false`，收藏量减少且不小于 0 | 通过 |
| TC-26 | 买家重新收藏商品 | POST | `/favorites` | 返回 `favorited=true`，收藏量增加 | 通过，`favorite_count` 由 0 增加为 1 |
| TC-27 | 检查收藏状态 | GET | `/favorites/check?productId=7` | 返回 `favorited=true` | 通过 |
| TC-28 | 收藏后管理员统计 | GET | `/stats/products/overview`、`/stats/products/category-distribution` | `totalViews`、`totalFavorites` 同步变化 | 通过，均返回 1 |

---

## 7. 浏览量与收藏量联动专项测试

### 7.1 测试商品

```text
商品ID：7
标题：浏览收藏统计测试商品
分类：电子产品
价格：120
成色：九成新
卖家用户ID：6
买家用户ID：8
```

### 7.2 浏览量测试

买家访问商品详情：

```bash
curl "http://127.0.0.1:8080/products/7" ^
-H "Authorization: Bearer %BUYER_TOKEN%"
```

访问后查询数据库：

```sql
SELECT id, title, view_count, favorite_count
FROM products
WHERE id = 7;
```

返回结果：

```text
+----+--------------------------------+------------+----------------+
| id | title                          | view_count | favorite_count |
+----+--------------------------------+------------+----------------+
|  7 | 浏览收藏统计测试商品           |          1 |              0 |
+----+--------------------------------+------------+----------------+
```

结论：浏览量修复通过，访问商品详情后 `view_count` 能够自动增加。

### 7.3 收藏量测试

因为测试前商品 7 已经存在旧收藏记录，因此先取消收藏，再重新收藏，以验证新代码逻辑。

#### 7.3.1 取消收藏

```bash
curl -X DELETE "http://127.0.0.1:8080/favorites/7" ^
-H "Authorization: Bearer %BUYER_TOKEN%"
```

返回：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "favorited": false
  },
  "timestamp": "2026-06-10 16:46:53"
}
```

#### 7.3.2 重新收藏

```bash
curl -X POST http://127.0.0.1:8080/favorites ^
-H "Content-Type: application/json" ^
-H "Authorization: Bearer %BUYER_TOKEN%" ^
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
  "timestamp": "2026-06-10 16:47:00"
}
```

#### 7.3.3 数据库验证

```sql
SELECT id, title, view_count, favorite_count
FROM products
WHERE id = 7;

SELECT * FROM favorites WHERE product_id = 7;
```

返回：

```text
+----+--------------------------------+------------+----------------+
| id | title                          | view_count | favorite_count |
+----+--------------------------------+------------+----------------+
|  7 | 浏览收藏统计测试商品           |          1 |              1 |
+----+--------------------------------+------------+----------------+

+----+---------+------------+---------------------+------------+
| id | user_id | product_id | create_time         | is_deleted |
+----+---------+------------+---------------------+------------+
|  1 |       8 |          7 | 2026-06-10 10:12:08 |          0 |
+----+---------+------------+---------------------+------------+
```

结论：收藏量修复通过。收藏关系有效，`favorites.is_deleted=0`，同时 `products.favorite_count` 已同步增加为 1。

### 7.4 管理员统计接口验证

#### 7.4.1 商品总览统计

请求：

```bash
curl "http://127.0.0.1:8080/stats/products/overview" ^
-H "Authorization: Bearer %ADMIN_TOKEN%"
```

返回：

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
    "totalViews": 1,
    "totalFavorites": 1,
    "averagePrice": 66.4
  },
  "timestamp": "2026-06-10 16:50:42"
}
```

结论：统计总览接口已同步显示浏览量和收藏量。

#### 7.4.2 分类分布统计

请求：

```bash
curl "http://127.0.0.1:8080/stats/products/category-distribution" ^
-H "Authorization: Bearer %ADMIN_TOKEN%"
```

返回：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "categoryId": 2,
      "categoryName": "电子产品",
      "productCount": 4,
      "onSaleCount": 2,
      "averagePrice": 135,
      "totalViews": 1,
      "totalFavorites": 1
    },
    {
      "categoryId": 1,
      "categoryName": "教材资料",
      "productCount": 3,
      "onSaleCount": 2,
      "averagePrice": 20.666667,
      "totalViews": 0,
      "totalFavorites": 0
    },
    {
      "categoryId": 3,
      "categoryName": "生活用品",
      "productCount": 0,
      "onSaleCount": 0,
      "averagePrice": 0,
      "totalViews": 0,
      "totalFavorites": 0
    }
  ],
  "timestamp": "2026-06-10 16:50:50"
}
```

结论：电子产品分类下 `totalViews=1`、`totalFavorites=1`，分类统计接口同步通过。

---

## 8. 其他核心功能测试记录

### 8.1 获取商品分类

请求：

```bash
curl "http://127.0.0.1:8080/categories"
```

结果：返回 6 个启用分类，包括教材资料、电子产品、生活用品、服饰鞋包、运动户外、其他。

结论：通过。

### 8.2 商品列表、搜索和筛选

测试内容：

```text
GET /products
GET /products?keyword=台灯
GET /products?categoryId=2
GET /products?minPrice=100&maxPrice=200
GET /products?keyword=台灯&categoryId=2&minPrice=100&maxPrice=200&conditionLevel=九成新&sort=price_asc
GET /products?keyword=台灯&categoryId=2&minPrice=100&maxPrice=200&conditionLevel=全新&sort=price_asc
```

结果：列表、关键词搜索、分类筛选、价格区间筛选、成色筛选均正常，空结果返回 `list:[]`、`total:0`。

结论：通过。

### 8.3 发布、编辑、上下架、删除商品

测试内容：

```text
POST /products
GET /products/my
PUT /products/:id
PUT /products/:id/status
DELETE /products/:id
```

结果：认证用户可正常发布商品、查看自己的商品、编辑商品、上下架商品和删除商品；未认证用户发布商品返回 `403 student verification required`。

结论：通过。

### 8.4 图片上传和绑定

测试内容：

```text
POST /upload/image
POST /products/:id/images
```

结果：图片上传后返回图片 URL，绑定商品图片后商品详情中可查询到图片列表。

结论：通过。

### 8.5 AI 商品优化入口

测试内容：

```text
POST /ai/optimize-product
```

结果：本地未配置 AI API Key 时返回：

```json
{
  "code": 500,
  "message": "AI服务暂不可用：未配置 API Key",
  "data": null
}
```

结论：接口入口正常，未配置 Key 时返回可控错误，不影响普通商品发布流程。

---

## 9. 已发现问题与处理记录

### 9.1 本地数据库缺少 `token_version` 字段

- 问题：合并最新代码后，登录接口报 `Unknown column 'token_version' in 'field list'`。
- 原因：本地数据库表结构未同步最新 `dev`。
- 处理：在 `users` 表中补充 `token_version` 字段。
- 结果：登录恢复正常。

### 9.2 未认证用户无法发布或收藏商品

- 问题：普通用户发布商品或收藏商品返回 `403 student verification required`。
- 原因：最新代码要求相关业务操作用户完成学生认证。
- 处理：将测试用户 `auth_status` 设置为 `VERIFIED` 后重新测试。
- 结果：认证用户可正常发布商品和收藏商品。

### 9.3 下架商品普通详情接口不可见

- 现象：商品状态改为 `OFF_SHELF` 后，普通详情接口返回 `product not found`。
- 判断：该行为符合当前业务逻辑，下架商品不对普通用户展示。
- 处理：记录在测试说明中，无需修改代码。

### 9.4 AI 未配置 API Key

- 现象：调用 AI 优化接口返回 `AI服务暂不可用：未配置 API Key`。
- 判断：本地测试环境未配置真实 AI API Key，属于可控错误。
- 处理：记录在测试说明中，不影响用户手动发布商品。

### 9.5 浏览量统计联动修复记录

- 原问题：买家访问 `GET /products/7` 商品详情后，商品详情中的 `viewCount` 仍为 0。
- 修复方式：在商品详情查询流程中增加 `products.view_count` 自增逻辑。
- 验证结果：访问商品详情后，`products.view_count` 由 0 增加为 1，管理员统计接口 `totalViews` 同步变为 1。

### 9.6 收藏量统计联动修复记录

- 原问题：买家收藏商品成功后，`favorites` 表存在有效收藏记录，但 `products.favorite_count` 未同步更新。
- 修复方式：将收藏关系写入与 `products.favorite_count` 更新放入同一事务中处理；收藏成功时 `favorite_count +1`，取消收藏时 `favorite_count -1`，并通过 `CASE` 保证不会小于 0。
- 验证结果：取消收藏后重新收藏，`favorites.is_deleted=0`，`products.favorite_count` 由 0 增加为 1；管理员统计接口 `totalFavorites` 同步变为 1。

---

## 10. 测试结论

在合并最新 `dev` 分支代码并完成本次修复后，商品相关模块编译通过，分类、商品列表、搜索筛选、成色筛选、商品详情、发布商品、我的商品、编辑商品、上下架、图片上传、图片绑定、AI 优化入口、商品统计权限、管理员统计接口和商品删除等核心功能均已完成验证。

本次重点修复的浏览量与收藏量统计联动问题已经通过验证：

1. 访问商品详情 `GET /products/:id` 后，`products.view_count` 能够自动增加；
2. 收藏商品 `POST /favorites` 成功后，`products.favorite_count` 能够自动增加；
3. 取消收藏 `DELETE /favorites/:productId` 后，`products.favorite_count` 能够自动减少且不会小于 0；
4. `favorites` 表与 `products.favorite_count` 能够保持一致；
5. 管理员统计接口 `/stats/products/overview` 和 `/stats/products/category-distribution` 中的 `totalViews`、`totalFavorites` 能够随商品浏览量和收藏量同步变化。

因此，商品模块浏览量、收藏量及统计接口联动修复通过。
