# outlookcli

在终端通过 [Microsoft Graph](https://learn.microsoft.com/en-us/graph/overview) 访问 Outlook 邮件与日历（需事先完成 OAuth，与 `~/.outlook-mcp` 下的 `config.json` / `credentials.json` 兼容）。

## 环境要求

- **Go**：1.21 及以上（Linux / macOS / Windows 均可编译）
- **账号配置**：`config.json` 中含 `client_id`、`client_secret`；`credentials.json` 中含 `refresh_token`（可用 Azure CLI 脚本或门户注册应用后授权生成）

## 安装方式

### 1. Homebrew（macOS / Linux 上装了 Homebrew 时）

Homebrew **不会**自动收录个人项目，需要自己建一个 **tap**（小型 formula 仓库），或在本机用本地 formula 安装。

#### 方式 A：自建 tap（适合团队 / 长期分发）

1. 在 GitHub 新建仓库，命名建议为 **`homebrew-outlookcli`**（`homebrew-` 前缀可让 tap 名更短）。
2. 在本仓库里复制 **`Formula/outlookcli.rb`** 到新仓库的 **`Formula/outlookcli.rb`**。
3. 用编辑器把 formula 里的 **`homepage`、`head` 的 GitHub 地址**改成 **存放 outlookcli 源码的仓库**（`main` 分支上有 `cmd/outlookcli` 的那个），不是 `homebrew-*` tap 仓库。
4. 推送 `main` 分支后，在本机执行：

```bash
brew tap ysu03zyy/outlookcli https://github.com/ysu03zyy/homebrew-outlookcli
brew update
brew install --HEAD outlookcli
```

说明：当前 formula 以 **`head`** 为主（从 `main` 源码用本机 Go 编译），因此需要加 **`--HEAD`**。若你希望 **`brew install outlookcli` 不带 `--HEAD`**，需要在发 **Git tag**（如 `v0.1.0`）后，在 formula 里增加 **`url` + `sha256`**（源码包地址一般为  
`https://github.com/ysu03zyy/outlookcli/archive/refs/tags/v0.1.0.tar.gz`），校验和可用：

```bash
curl -sL "https://github.com/ysu03zyy/outlookcli/archive/refs/tags/v0.1.0.tar.gz" | shasum -a 256
```

把输出填进 formula 的 `sha256` 字段，并取消注释 `url` / `sha256` 两行（见 `Formula/outlookcli.rb` 内注释）。

#### 方式 B：不建 tap，直接用本仓库里的 formula（本地开发）

在克隆了 **outlookcli 源码**的机器上：

```bash
cd outlookcli
brew install --HEAD --build-from-source ./Formula/outlookcli.rb
```

同样需要本机已安装 **Go**（`depends_on "go" => :build`）。

---

### 2. 从本仓库目录安装（推荐本地开发）

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

### 3. 指定版本号编译（可选）

```bash
cd outlookcli
go build -ldflags "-X main.version=1.0.0" -o outlookcli ./cmd/outlookcli
sudo mv outlookcli /usr/local/bin/   # 或放到任意在 PATH 中的目录
```

### 4. 使用 Makefile

```bash
cd outlookcli
make install      # go install ./cmd/outlookcli
make build        # 输出到 bin/outlookcli
```

交叉编译示例（在 macOS 上生成 Linux 二进制）：

```bash
make release-all   # 生成 bin/ 下多架构二进制，见 Makefile
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
