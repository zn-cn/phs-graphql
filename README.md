# 写了吗小程序后台 graphql版本

注：REST版本见

技术栈：golang, graphql, jwt, mongo, redis

命名：所有字段全部采用驼峰命名（含数据库和返回字段）

## 项目介绍

### 简介

针对班级作业通知的小公举，确保通知到每一位同学，作业写了吗

#### 认证：

过期时间：七天

- JWT 头部字段：Authorization ，需要前端手动在token前面加上 Bearer加空格

##### JWT认证流程：

- 前端使用 `wx.login` 拿到 微信的code
- 再使用 `wx.getUserInfo` 获取微信信息
- 调用后台 API：POST `/api/v1/login`

## 附录

### docker-compose 说明

+ 环境变量
  + CONFIG_PATH_PREFIX  配置文件路径前缀
  + ENV 环境，同时对应配置文件名字，如: prod  -> prod.json
  + QINIU_ACCESS_KEY 七牛云 access_key
  + QINIU_SECRET_KEY 七牛云 secret_key
  + QINIU_BUCKET 七牛云空间名字
+ 数据库

### 图片

存储：七牛云存储

区域：华南

bucket：php-mp

区域域名：https://upload-z2.qiniup.com

[七牛云存储上传文档](https://developer.qiniu.com/kodo/manual/1272/form-upload)

注：后台设置了图片上传时自动生成缩略图

+ 微信默认头像：<wechat-default-headimgurl.jpg>

+ 图片前缀

  域名：`<image host>`


参考链接：

- [graphql-go](https://github.com/graphql-go/graphql)
- [graphql-go-handler](https://github.com/graphql-go/handler)
- [graphql](https://graphql.org/learn/queries/)

  ​