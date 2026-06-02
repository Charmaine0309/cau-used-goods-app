# 前端A微信小程序运行说明

这个目录是按照 Git 仓库远端 dev 分支的真实接口调整后的微信开发者工具可运行版本。

## 打开方式

微信开发者工具 -> 导入项目 -> 选择本目录：

C:\Users\14770\Desktop\cau-used-goods-frontend-A-fixed

## 重要：先同步后端

远端 dev 已经新增商品、收藏、订单、举报等接口。你本地旧后端如果没同步，会没有这些接口。

建议在 Git 仓库中同步 dev 后再启动后端：

cd C:\Users\14770\Desktop\cau-used-goods-app
git fetch origin
git checkout dev
git pull origin dev
cd backend
go run .\cmd\server

看到 server listening on :8080 后，再在微信开发者工具里编译前端。

## 已对接真实接口

- POST /auth/dev-login：登录并保存 token。
- GET /users/me：个人中心获取当前用户资料。
- PUT /users/profile：修改昵称、手机号、头像地址。
- POST /users/avatar：上传头像。
- POST /users/student-verify：提交学生认证。
- GET /users/student-verify：查看学生认证状态。
- GET /admin/users/student-verifications：管理员查看待审核学生认证。
- PUT /admin/users/:id/student-verify：管理员审核学生认证。
- GET /stats/users/overview：用户数据看板。
- GET /stats/products/overview：商品数据看板。
- GET /stats/orders/overview：订单数据看板。
- GET /admin/logs：管理员日志。
- GET /categories：首页、搜索、分类页加载商品分类。
- GET /products：首页最新商品、搜索结果、分类商品列表。
- GET /products/:id：商品详情。
- POST /favorites：收藏商品。
- DELETE /favorites/:productId：取消收藏。
- GET /favorites/check：检查收藏状态。
- POST /orders：提交预约。
- POST /reports：举报商品。

## 注意

后端 SQL 当前只初始化分类，不初始化商品。如果数据库 products 表没有商品数据，首页和搜索页会显示暂无商品。这种情况需要先由发布商品模块写入商品数据，或者由后端同学提供测试商品数据。

地址管理当前为前端本地缓存实现，后端提供地址接口后再替换为真实接口。
