# MyBlog

MyBlog 是一个 Gitea 风格的自托管知识库/博客系统骨架：单个 Go exe 启动，读取 `app.ini`，首次通过 Web 安装页初始化数据库和管理员账号，然后提供用户、团队、空间、文章、附件和全文搜索能力。

## 技术栈

- Go 后端：`cmd -> routers -> services -> models -> modules`
- HTTP 路由：chi
- 数据层：XORM，支持 SQLite、MySQL、PostgreSQL
- 前端：Go `html/template` 渲染页面，Vue 3 + TypeScript + Vite 源码位于 `web_src`
- 搜索：内嵌 Bleve，索引落在 `data/indexers/articles`
- 内容：Markdown 为标准存储格式，渲染后经 sanitizer 清洗

## 个人功能

- `/users/{name}`：公开个人主页，展示个人资料和可见的个人博文。
- `/user/settings`：登录后维护显示名称、邮箱、头像、网站、所在地、简介和密码。
- `/user/articles/new`：登录后直接写个人博文。系统会自动创建内部个人团队 `user-<id>` 和 `blog` 空间，不需要用户先手动创建团队。

## 启动

```powershell
go run . web -c custom/conf/app.ini --work-path .
```

首次访问会跳转到 `/install`。默认 SQLite 数据库路径为 `data/myblog.db`，安装完成后会生成 `custom/conf/app.ini` 并设置 `INSTALL_LOCK=true`。

## 预留命令

```powershell
myblog admin
myblog migrate
myblog doctor
myblog indexer rebuild
```

这些命令入口已经保留，后续可以继续实现具体运维行为。
