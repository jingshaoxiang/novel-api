# NovelAPI

💥 支持 AccessToken 使用账号画图。

🔍 回复格式与真实 API 完全一致，适配几乎所有客户端

🔍 图片存放Alist平台，免存储！

### 逆向API 功能
> - [x] 对话式画图

### 改造功能
> - [x] 图片存放至Alist平台。
> - [x] 后台账号池随机抽取。
> - [x] 错误账号自动剔除禁用。
> - [x] /web 可直接进入Token管理页面。

> TODO
> - [ ] 暂无，欢迎提 `issue`

```bash
curl --location 'http://127.0.0.1:3388/v1/chat/completions' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer {{Token}}' \
--data '{
     "model": "nai-diffusion-3",
     "messages": [{"role": "user", "content": "Say this is a test!"}],
     "stream": true
   }'
```

## Tokens 管理

1. 访问 `/web` ， 可以查看现有 Tokens 数量，也可以上传新的 Tokens ，或者清空 Tokens。
![img.png](img.png)

## 部署

### 部署AList
[安装Alist组件](https://www.master-jsx.top/archives/alistpan)

### 直接部署

```bash
git clone https://github.com/jingshaoxiang/novel-api.git
cd novel-api
# 填写config.yml中的环境变量，然后运行以下命令启动应用程序。
./novel-x86
```

### Docker 部署

您需要安装 Docker 和 Docker Compose。

```bash
# 请根据您自己的系统类型导入最新的包

docker load < ./docker/novel-x86.tar

docker-compose up -d # 启动容器
```

### 密钥获取方式
1.访问 https://novelai.net/image 创建账号，打开 F12 查看，找到 负载 中有数据的接口。
![img_1.png](img_1.png)
![img_2.png](img_2.png)


### 进入到 http://127.0.0.1:3388/web 填入Key即可使用！
![img_3.png](img_3.png)

### 进入到New-api中添加渠道
![img_4.png](img_4.png)

### 即可正常调用画图
```azure
### 画图格式为：
正词：想画的关键词 
    
反词：不想画出来的关键词
```

### 效果
![img_5.png](img_5.png)

### 如果不符合格式则会出来白毛
![img_6.png](img_6.png)