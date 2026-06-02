# 前端A账户与系统管理模块说明

负责人：潘欣怡

## 模块范围

按照《前端开发说明.MD》的新分工，前端A负责账户与系统管理相关页面：

- 登录入口
- 学生认证
- 个人中心
- 资料修改
- 地址管理
- 管理员后台入口
- 数据看板与学生认证审核

## 页面清单

| 页面 | 路径 | 说明 |
| --- | --- | --- |
| 登录页 | `pages/login/login` | 统一登录入口，登录后根据后端返回角色判断用户身份 |
| 个人中心 | `pages/index/index` | 展示头像、昵称、学生认证状态，进入资料修改、学生认证、地址管理和后台管理 |
| 修改资料 | `pages/profile-edit/profile-edit` | 修改昵称、手机号，上传头像，可选择沿用微信昵称和头像 |
| 学生认证 | `pages/student-auth/student-auth` | 填写姓名、学号、学院并提交审核 |
| 地址管理 | `pages/address/address` | 新增、删除、设置默认交易地址 |
| 后台管理 | `pages/admin/admin` | 查看数据看板、待审核学生认证、管理员日志 |

## 跳转逻辑

1. 用户进入小程序后先进入登录页。
2. 登录成功后进入首页，首页右上角可进入个人中心。
3. 个人中心中可进入学生认证、资料修改、地址管理。
4. 当后端返回用户 `role=ADMIN` 时，个人中心显示“后台管理”入口；登录页不单独暴露管理员入口。
5. 管理员进入后台管理后，可以查看统计数据和待审核学生认证。

## 接口对接

| 功能 | 接口 |
| --- | --- |
| 开发环境登录 | `POST /auth/dev-login` |
| 获取当前用户 | `GET /users/me` |
| 修改资料 | `PUT /users/profile` |
| 上传头像 | `POST /users/avatar` |
| 提交学生认证 | `POST /users/student-verify` |
| 查看学生认证 | `GET /users/student-verify` |
| 管理员查看待审核认证 | `GET /admin/users/student-verifications` |
| 管理员审核认证 | `PUT /admin/users/:id/student-verify` |
| 用户统计 | `GET /stats/users/overview` |
| 商品统计 | `GET /stats/products/overview` |
| 订单统计 | `GET /stats/orders/overview` |
| 管理员日志 | `GET /admin/logs` |

## 说明

地址管理当前使用本地缓存实现，原因是后端当前 Git 版本未提供地址相关接口。后续如果后端增加地址接口，可将 `api/address.js` 从本地缓存替换为真实请求。

管理员后台入口不面向普通用户展示，只有登录用户的 `role` 为 `ADMIN` 时才显示。管理员身份由后端数据库中的用户角色决定，前端不通过按钮决定管理员权限。
