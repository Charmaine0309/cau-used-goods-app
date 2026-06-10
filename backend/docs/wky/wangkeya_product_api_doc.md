# 王珂雅商品相关模块接口文档（最终版）

## 1. 文档说明

本文档用于说明 CAU 校园二手交易平台中王珂雅负责的商品相关后端模块接口，覆盖商品分类、商品列表、搜索筛选、商品详情、商品发布、商品编辑、商品上下架、商品删除、商品图片上传与绑定、AI 商品优化、商品统计以及收藏/浏览统计联动检查等内容。

- 模块负责人：王珂雅
- 后端基础地址：`http://127.0.0.1:8080`
- 数据格式：JSON
- 鉴权方式：需要登录的接口通过请求头携带 JWT Token

```http
Authorization: Bearer <token>
```

---

## 2. 通用返回格式

### 2.1 成功返回

```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "timestamp": "2026-06-10 10:12:08"
}
```

### 2.2 失败返回

```json
{
  "code": 403,
  "message": "需要管理员权限",
  "data": null,
  "timestamp": "2026-06-10 10:08:12"
}
```

---

## 3. 商品分类接口

### 3.1 获取商品分类列表

- 请求方式：`GET`
- 接口路径：`/categories`
- 是否需要登录：否
- 功能说明：获取平台启用的商品分类列表。

#### 请求示例

```bash
curl "http://127.0.0.1:8080/categories"
```

#### 返回示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {"id":1,"name":"教材资料","parentId":0,"sortOrder":10,"status":"ENABLED"},
    {"id":2,"name":"电子产品","parentId":0,"sortOrder":20,"status":"ENABLED"},
    {"id":3,"name":"生活用品","parentId":0,"sortOrder":30,"status":"ENABLED"},
    {"id":4,"name":"服饰鞋包","parentId":0,"sortOrder":40,"status":"ENABLED"},
    {"id":5,"name":"运动户外","parentId":0,"sortOrder":50,"status":"ENABLED"},
    {"id":6,"name":"其他","parentId":0,"sortOrder":999,"status":"ENABLED"}
  ],
  "timestamp": "2026-06-09 18:33:04"
}
```

---

## 4. 商品列表、搜索与筛选接口

### 4.1 获取商品列表

- 请求方式：`GET`
- 接口路径：`/products`
- 是否需要登录：否
- 功能说明：分页获取在售商品列表，支持关键词、分类、价格区间、成色和排序筛选。

#### 查询参数

| 参数名 | 类型 | 是否必填 | 说明 |
|---|---|---|---|
| `keyword` | string | 否 | 关键词搜索，匹配商品标题或描述 |
| `categoryId` | int | 否 | 商品分类 ID |
| `minPrice` | number | 否 | 最低价格 |
| `maxPrice` | number | 否 | 最高价格 |
| `conditionLevel` | string | 否 | 商品成色，按 `condition_level` 精确匹配 |
| `sort` | string | 否 | 排序方式，如 `price_asc`、`price_desc`、`popular` |
| `page` | int | 否 | 页码，默认 1 |
| `pageSize` | int | 否 | 每页数量，默认 10 |

#### 请求示例

```bash
curl "http://127.0.0.1:8080/products"

curl "http://127.0.0.1:8080/products?keyword=台灯"

curl "http://127.0.0.1:8080/products?keyword=台灯&categoryId=2&minPrice=100&maxPrice=200&conditionLevel=九成新&sort=price_asc"
```

#### 返回示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 5,
        "sellerId": 5,
        "categoryId": 2,
        "title": "宿舍自用台灯",
        "description": "九成新台灯，适合宿舍学习使用",
        "originalPrice": 199,
        "price": 150,
        "conditionLevel": "九成新",
        "meetLocation": "图书馆门口",
        "status": "ON_SALE",
        "viewCount": 0,
        "favoriteCount": 0,
        "createTime": "2026-06-03 10:33:26",
        "images": []
      }
    ],
    "page": 1,
    "pageSize": 10,
    "total": 1
  },
  "timestamp": "2026-06-09 18:33:41"
}
```

#### 空结果返回示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [],
    "page": 1,
    "pageSize": 10,
    "total": 0
  },
  "timestamp": "2026-06-09 18:33:46"
}
```

### 4.2 成色筛选说明

`conditionLevel` 当前采用精确匹配逻辑，后端 SQL 条件为：

```sql
AND condition_level = ?
```

因此：

- 数据库存储 `九成新`，前端传 `conditionLevel=九成新`，可以筛出；
- 数据库存储 `九成新，非常不错`，前端传 `conditionLevel=九成新`，不会筛出。

建议前端将成色字段设计为固定选项，例如：`全新`、`九成新`、`八成新`、`七成新`、`六成新及以下`。用户补充说明如“非常不错”“几乎没用过”等应填写到 `description` 字段。

---

## 5. 商品详情接口

### 5.1 获取商品详情

- 请求方式：`GET`
- 接口路径：`/products/:id`
- 是否需要登录：否
- 功能说明：根据商品 ID 获取商品详情和图片列表。

#### 请求示例

```bash
curl "http://127.0.0.1:8080/products/7"
```

#### 返回示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 7,
    "sellerId": 6,
    "categoryId": 2,
    "title": "浏览收藏统计测试商品",
    "description": "用于测试浏览量和收藏量统计",
    "originalPrice": 199,
    "price": 120,
    "conditionLevel": "九成新",
    "meetLocation": "图书馆门口",
    "status": "ON_SALE",
    "viewCount": 0,
    "favoriteCount": 0,
    "createTime": "2026-06-10 10:06:00",
    "images": []
  },
  "timestamp": "2026-06-10 10:12:20"
}
```

### 5.2 下架或删除后的详情表现

商品下架或删除后，普通详情接口可能返回：

```json
{
  "code": 404,
  "message": "product not found",
  "data": null
}
```

该行为表示商品不再对普通用户展示。

---

## 6. 商品发布接口

### 6.1 发布商品

- 请求方式：`POST`
- 接口路径：`/products`
- 是否需要登录：是
- 权限要求：用户需完成学生认证，即 `authStatus=VERIFIED`
- 功能说明：发布新的二手商品。

#### 请求头

```http
Content-Type: application/json
Authorization: Bearer <token>
```

#### 请求参数

| 参数名 | 类型 | 是否必填 | 说明 |
|---|---|---|---|
| `categoryId` | int | 是 | 商品分类 ID |
| `title` | string | 是 | 商品标题 |
| `description` | string | 是 | 商品描述 |
| `originalPrice` | number | 否 | 原价 |
| `price` | number | 是 | 售价 |
| `conditionLevel` | string | 是 | 商品成色 |
| `meetLocation` | string | 是 | 交易地点 |

#### 请求示例

```bash
curl -X POST http://127.0.0.1:8080/products ^
-H "Content-Type: application/json" ^
-H "Authorization: Bearer <token>" ^
-d "{"categoryId":2,"title":"浏览收藏统计测试商品","description":"用于测试浏览量和收藏量统计","originalPrice":199,"price":120,"conditionLevel":"九成新","meetLocation":"图书馆门口"}"
```

#### 返回示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 7
  },
  "timestamp": "2026-06-10 10:06:00"
}
```

#### 未完成学生认证时返回

```json
{
  "code": 403,
  "message": "student verification required",
  "data": null
}
```

---

## 7. 我的商品接口

### 7.1 获取当前用户发布的商品

- 请求方式：`GET`
- 接口路径：`/products/my`
- 是否需要登录：是
- 功能说明：获取当前登录用户发布的商品列表。

#### 请求示例

```bash
curl "http://127.0.0.1:8080/products/my" ^
-H "Authorization: Bearer <token>"
```

---

## 8. 商品编辑接口

### 8.1 编辑商品

- 请求方式：`PUT`
- 接口路径：`/products/:id`
- 是否需要登录：是
- 权限要求：只能编辑自己发布的商品
- 功能说明：修改商品标题、描述、价格、成色和交易地点等信息。

#### 请求示例

```bash
curl -X PUT http://127.0.0.1:8080/products/6 ^
-H "Content-Type: application/json" ^
-H "Authorization: Bearer <token>" ^
-d "{"categoryId":2,"title":"宿舍自用台灯测试版-已修改","description":"修改后的九成新台灯，功能正常","originalPrice":199,"price":140,"conditionLevel":"九成新","meetLocation":"东区食堂门口"}"
```

#### 返回示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 6
  }
}
```

---

## 9. 商品状态接口

### 9.1 修改商品状态

- 请求方式：`PUT`
- 接口路径：`/products/:id/status`
- 是否需要登录：是
- 权限要求：商品发布者或具备管理权限的用户
- 功能说明：修改商品状态，例如上架、下架等。

#### 状态值说明

| 状态值 | 说明 |
|---|---|
| `ON_SALE` | 在售 |
| `OFF_SHELF` | 已下架 |
| `SOLD` | 已售出 |
| `LOCKED` | 锁定中 |
| `DELETED` | 已删除 |

#### 请求示例

```bash
curl -X PUT http://127.0.0.1:8080/products/6/status ^
-H "Content-Type: application/json" ^
-H "Authorization: Bearer <token>" ^
-d "{"status":"OFF_SHELF"}"
```

#### 返回示例

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

---

## 10. 图片上传与绑定接口

### 10.1 上传商品图片

- 请求方式：`POST`
- 接口路径：`/upload/image`
- 是否需要登录：是
- 请求类型：`multipart/form-data`
- 功能说明：上传商品图片，并返回图片访问路径。

#### 请求示例

```bash
curl -X POST http://127.0.0.1:8080/upload/image ^
-H "Authorization: Bearer <token>" ^
-F "file=@C:\Users\18472\Desktop\test.png"
```

#### 返回示例

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

### 10.2 绑定商品图片

- 请求方式：`POST`
- 接口路径：`/products/:id/images`
- 是否需要登录：是
- 权限要求：只能为自己发布的商品绑定图片
- 功能说明：将图片 URL 绑定到指定商品。

#### 请求示例

```bash
curl -X POST http://127.0.0.1:8080/products/6/images ^
-H "Content-Type: application/json" ^
-H "Authorization: Bearer <token>" ^
-d "{"images":["/uploads/products/1781001668926493800.png"]}"
```

#### 返回示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 6,
    "images": [
      "/uploads/products/1781001668926493800.png"
    ]
  },
  "timestamp": "2026-06-09 18:42:29"
}
```

---

## 11. AI 商品优化接口

### 11.1 AI 优化商品标题和描述

- 请求方式：`POST`
- 接口路径：`/ai/optimize-product`
- 是否需要登录：是
- 功能说明：根据商品标题和描述调用 AI 服务进行优化。

#### 请求示例

```bash
curl -X POST http://127.0.0.1:8080/ai/optimize-product ^
-H "Content-Type: application/json" ^
-H "Authorization: Bearer <token>" ^
-d "{"title":"高数教材","description":"正版教材，九成新，适合期末复习"}"
```

#### 未配置 API Key 时返回示例

```json
{
  "code": 500,
  "message": "AI服务暂不可用：未配置 API Key",
  "data": null,
  "timestamp": "2026-06-09 18:38:51"
}
```

该返回属于本地测试环境下的可控错误，表示 AI 服务入口可访问，但未配置外部 AI 服务密钥。

---

## 12. 收藏接口及商品收藏量联动说明

### 12.1 添加收藏

- 请求方式：`POST`
- 接口路径：`/favorites`
- 是否需要登录：是
- 权限要求：用户需完成学生认证
- 功能说明：收藏指定商品。

#### 请求示例

```bash
curl -X POST http://127.0.0.1:8080/favorites ^
-H "Content-Type: application/json" ^
-H "Authorization: Bearer <buyer_token>" ^
-d "{"productId":7}"
```

#### 返回示例

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

### 12.2 检查收藏状态

- 请求方式：`GET`
- 接口路径：`/favorites/check?productId=:id`
- 是否需要登录：是
- 功能说明：检查当前用户是否已收藏某商品。

#### 请求示例

```bash
curl "http://127.0.0.1:8080/favorites/check?productId=7" ^
-H "Authorization: Bearer <buyer_token>"
```

#### 返回示例

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

### 12.3 当前联动检查结果

测试发现：收藏接口本身可以成功创建收藏关系，但商品详情中的 `favoriteCount` 未同步增加，管理员统计接口中的 `totalFavorites` 也未同步更新。

数据库验证结果：

```sql
SELECT * FROM favorites WHERE product_id = 7;
```

返回：

```text
id=1, user_id=8, product_id=7, is_deleted=0
```

但：

```sql
SELECT id, title, view_count, favorite_count
FROM products
WHERE id = 7;
```

返回：

```text
id=7, view_count=0, favorite_count=0
```

说明收藏关系已创建，但 `products.favorite_count` 尚未与收藏流程完成联动更新。

---

## 13. 商品统计接口

商品统计接口需要管理员权限，普通用户访问会返回 `403`。

### 13.1 商品总览统计

- 请求方式：`GET`
- 接口路径：`/stats/products/overview`
- 是否需要登录：是
- 权限要求：管理员

#### 请求示例

```bash
curl "http://127.0.0.1:8080/stats/products/overview" ^
-H "Authorization: Bearer <admin_token>"
```

#### 返回示例

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

### 13.2 分类分布统计

- 请求方式：`GET`
- 接口路径：`/stats/products/category-distribution`
- 是否需要登录：是
- 权限要求：管理员

#### 返回字段说明

| 字段 | 说明 |
|---|---|
| `categoryId` | 分类 ID |
| `categoryName` | 分类名称 |
| `productCount` | 商品总数 |
| `onSaleCount` | 在售商品数 |
| `averagePrice` | 平均价格 |
| `totalViews` | 总浏览量 |
| `totalFavorites` | 总收藏量 |

#### 返回示例

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
      "totalViews": 0,
      "totalFavorites": 0
    }
  ],
  "timestamp": "2026-06-10 10:12:47"
}
```

### 13.3 状态分布统计

- 请求方式：`GET`
- 接口路径：`/stats/products/status-distribution`
- 是否需要登录：是
- 权限要求：管理员

### 13.4 商品发布趋势统计

- 请求方式：`GET`
- 接口路径：`/stats/products/trend?days=7`
- 是否需要登录：是
- 权限要求：管理员

---

## 14. 删除商品接口

### 14.1 删除商品

- 请求方式：`DELETE`
- 接口路径：`/products/:id`
- 是否需要登录：是
- 权限要求：只能删除自己发布的商品
- 功能说明：删除商品后，普通详情接口和商品列表不再展示该商品。

#### 请求示例

```bash
curl -X DELETE http://127.0.0.1:8080/products/6 ^
-H "Authorization: Bearer <token>"
```

#### 返回示例

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

#### 删除后详情返回示例

```json
{
  "code": 404,
  "message": "product not found",
  "data": null,
  "timestamp": "2026-06-09 18:47:08"
}
```

---

## 15. 与订单模块的商品状态联动说明

商品模块提供商品状态字段，订单模块在创建订单、成交、取消等流程中可联动更新商品状态。例如：

- 下单后可将商品状态改为 `LOCKED`；
- 交易完成后可将商品状态改为 `SOLD`；
- 订单取消后可恢复为 `ON_SALE`；
- 商品删除后状态为 `DELETED`，不再在普通商品列表中展示。

该部分与订单模块共同完成业务闭环。当前测试主要覆盖商品模块本身，不直接修改订单模块逻辑。
