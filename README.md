# outlookcli

在终端通过 [Microsoft Graph](https://learn.microsoft.com/en-us/graph/overview) 访问 Outlook 邮件与日历（需事先完成 OAuth，与 `~/.outlook-mcp` 下的 `config.json` / `credentials.json` 兼容）。

## 环境要求

- **Go**：1.21 及以上（Linux / macOS / Windows 均可编译）
- **账号配置**：`config.json` 中含 `client_id`、`client_secret`；`credentials.json` 中含 `refresh_token`（可用 Azure CLI 脚本或门户注册应用后授权生成）

## 安装方式

### 1. Linux / macOS 一键安装脚本（推荐）

发版后可直接执行：

```bash
curl -fsSL https://raw.githubusercontent.com/ysu03zyy/outlookcli/main/install.sh | bash
```

脚本会自动：

- 识别平台（`darwin/linux`）和架构（`amd64/arm64`）
- 下载对应 release 资产：`outlookcli_<version>_<os>_<arch>.tar.gz`
- 校验 `checksums.txt`
- 安装到 `/usr/local/bin/outlookcli`

可选环境变量：

```bash
# 安装指定版本（默认 latest）
OUTLOOKCLI_VERSION=0.1.0 curl -fsSL https://raw.githubusercontent.com/ysu03zyy/outlookcli/main/install.sh | bash

# 安装到自定义目录
OUTLOOKCLI_INSTALL_DIR="$HOME/.local/bin" curl -fsSL https://raw.githubusercontent.com/ysu03zyy/outlookcli/main/install.sh | bash
```

---

### 2. Homebrew（macOS 推荐）

项目不会自动进入 Homebrew core，建议维护一个 tap：

1. 新建仓库（例如 `homebrew-outlookcli`）
2. 复制本仓库 `Formula/outlookcli.rb` 到 tap 仓库 `Formula/outlookcli.rb`
3. 按最新 release 的 `checksums.txt` 更新 formula 里的 `version` 与各平台 `sha256`
4. 用户安装：

```bash
brew tap ysu03zyy/outlookcli https://github.com/ysu03zyy/homebrew-outlookcli
brew install outlookcli
```

---

### 3. 从本仓库目录安装（推荐本地开发）

模块路径为 **`github.com/ysu03zyy/outlookcli`**。在 `outlookcli` 目录下执行：

```bash
cd outlookcli
go mod tidy
go install ./cmd/outlookcli
```

若仓库已推送到 GitHub 且为公开仓库，也可在任意目录执行：

```bash
go install github.com/ysu03zyy/outlookcli/cmd/outlookcli@latest
```

安装后二进制一般在 **`$(go env GOPATH)/bin`**（例如 `~/go/bin/outlookcli`）。请确保该目录在 **`PATH`** 中：

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
# 可写入 ~/.zshrc 或 ~/.bashrc
```

验证：

```bash
outlookcli --help
outlookcli --version
```

### 4. 指定版本号编译（可选）

```bash
cd outlookcli
go build -ldflags "-X main.version=1.0.0" -o outlookcli ./cmd/outlookcli
sudo mv outlookcli /usr/local/bin/   # 或放到任意在 PATH 中的目录
```

### 5. 使用 Makefile

```bash
cd outlookcli
make install      # go install ./cmd/outlookcli
make build        # 输出到 bin/outlookcli
```

交叉编译示例（在 macOS 上生成 Linux 二进制）：

```bash
make release-all   # 生成 bin/ 下多架构二进制
```

---

### 6. 发布流程（给维护者）

仓库已内置 `.github/workflows/release.yml`。每次推送 tag（如 `v0.1.0`）会自动：

1. 构建四个平台资产（darwin/linux + amd64/arm64）
2. 打包成 `outlookcli_<version>_<os>_<arch>.tar.gz`
3. 生成 `checksums.txt`
4. 上传到 GitHub Release

发布命令示例：

```bash
git tag v0.1.0
git push origin v0.1.0
```

## 配置路径

- 默认目录：`~/.outlook-mcp`
- 或通过环境变量 / 参数覆盖：`OUTLOOK_CONFIG_DIR`、`--config-dir`
- 日历默认时区：`OUTLOOK_TIMEZONE` 或 `--timezone`（IANA，如 `Asia/Shanghai`）

### Access token 自动刷新（配置文件模式）

使用 `config.json` + `credentials.json` 时，`outlookcli` 通过 **`golang.org/x/oauth2` 的 `TokenSource`** 按需刷新（与 `gogcli` 的思路类似）：在 access token **未过期**时复用磁盘/内存中的 token，**过期后**再用 `refresh_token` 换新，并把新 token 及 **`expiry`** 写回 `credentials.json`。同一进程内多次 Graph 请求共用一个 `TokenSource`，不会在每次请求时都打 token 端点。

`outlookcli token refresh` 仍会**强制**走一轮刷新（与旧版 shell 脚本里“手动刷新”一致）。

## 多用户 / 直接传入 Access Token

不依赖本地 `config.json`、`credentials.json` 时，可用 **环境变量**或**全局参数**直接指定 Microsoft Graph 的 access token（适合多用户、上层系统代发 token、CI 等）：

- **`OUTLOOK_ACCESS_TOKEN`** 或 **`--access-token <token>`**
- 此模式下**不会**读取凭据文件，也**不会**用 refresh_token 刷新；token 过期（通常约 1 小时）后需自行换新或由调用方重新传入。
- **`outlookcli token refresh`** 仅适用于配置文件模式；若带了 `--access-token` / `OUTLOOK_ACCESS_TOKEN`，会报错并提示去掉直连 token。
- **`outlookcli token get`**：直连 token 模式下会把当前使用的 token 原样打印（即你传入的值）。

安全提示：在命令行里写 token 可能被同机用户通过 `ps` 看到；更稳妥的做法是用环境变量注入，或由进程管理器从密钥系统读取后设置 `OUTLOOK_ACCESS_TOKEN`。

示例：

```bash
export OUTLOOK_ACCESS_TOKEN="eyJ0eX..."
outlookcli mail inbox -n 5

outlookcli --access-token "$TOKEN" calendar today -z Asia/Shanghai
```

## 常用命令示例

```bash
outlookcli token test
outlookcli mail inbox -n 5
outlookcli mail read <id后缀>
outlookcli calendar today --timezone Asia/Shanghai
```

加 `-j` / `--json` 可输出 JSON，便于脚本解析。

## 与 Outlook skill 脚本的关系

使用同一套 Graph 与 `~/.outlook-mcp` 凭据目录；CLI 为单二进制，无需 `bash`/`jq`/`curl`。与旧脚本「每次调用先 refresh」不同，配置文件模式下 **`outlookcli` 仅在需要时刷新 token**（见上文「Access token 自动刷新」）。
