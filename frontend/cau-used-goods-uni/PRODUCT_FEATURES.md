# 商品模块说明

本目录基于 `feature/frontend-A-login-home-detail-v2` 的 uni-app 源码继续开发。`unpackage/` 是编译产物，已由 `.gitignore` 排除，不应作为业务源码修改。

## 微信开发者工具运行

使用 HBuilderX 打开 `frontend/cau-used-goods-uni`，执行：

```bash
运行 -> 运行到小程序模拟器 -> 微信开发者工具
```

然后在微信开发者工具中重新编译。没有 HBuilderX 时，可临时导入仓库中已有的编译产物：

```text
frontend/cau-used-goods-uni/unpackage/dist/dev/mp-weixin
```

注意：`unpackage/` 是旧编译产物，不作为协作修改重点；源码改动后必须用 HBuilderX 重新编译。

## 已实现功能

- 首页：分类入口、最新商品、分页加载、下拉刷新。
- 分类与搜索：关键词、分类、价格区间、成色、排序、分页和空状态。
- 商品详情：图片轮播、商品状态、卖家脱敏信息、收藏、预约和举报入口。
- 发布商品：表单校验、最多 9 张图片、单图 5 MB 限制、上传、删除图片。
- AI 辅助：标题优化候选、描述生成。AI 接口失败不会阻止用户手动发布。
- 权限入口：发布、收藏、预约和举报操作会检查登录与学生认证状态。

## 商品相关接口

```text
GET    /categories
GET    /products
GET    /products/:id
POST   /products
POST   /upload/image
POST   /ai/optimize-product
POST   /favorites
DELETE /favorites/:productId
GET    /favorites/check?productId=:productId
POST   /orders
POST   /reports
```

列表查询参数：

```text
keyword categoryId minPrice maxPrice conditionLevel sort page pageSize
```

图片上传接口预期返回：

```json
{
  "code": 0,
  "data": {
    "url": "/uploads/products/example.jpg"
  }
}
```
