# SSH 证书导入教程

## 1. 解压
下载的 zip 解压后得到私钥和用户证书两个文件。

## 2. 放置私钥和证书
- 建议放到 `~/.ssh/` 目录。
- 设置权限：`chmod 600 <私钥文件名>`。
- 证书和私钥必须在一个目录，证书文件名必须为 `{私钥文件名}-cert.pub`

## 3. 配置 ssh_config
在 `~/.ssh/config` 追加条目：
```
Host your-host
  HostName <服务器地址>
  User <登录用户名>
  IdentityFile ~/.ssh/<私钥文件名>
```

## 4. 测试连接
`ssh your-host -vv`

## 5. 相关链接
VSCode: [Using VSCode over SSH](https://technotes.adelerhof.eu/devops/vscode_over_ssh/)
Termus: [Import SSH Certificate](https://termius.com/documentation/ssh-certificates)

> 提示：证书有有效期，请在到期前重新申请。妥善保存私钥，勿泄露。