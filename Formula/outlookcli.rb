# Homebrew formula for outlookcli.
#
# 用法见仓库 README「Homebrew 安装」一节。
#
# 仅 HEAD（从 GitHub main 构建，无需发版）：
#   brew install --HEAD outlookcli
#   （取决于你的 tap 仓库名，见 README）
#
# 稳定版：取消下面 url / sha256 的注释，在发 tag 后把 sha256 换成真实校验和：
#   curl -sL "https://github.com/OWNER/outlookcli/archive/refs/tags/v0.1.0.tar.gz" | shasum -a 256

class Outlookcli < Formula
  desc "Outlook mail and calendar via Microsoft Graph"
  homepage "https://github.com/ysu03zyy/outlookcli"
  license "MIT"
  head "https://github.com/ysu03zyy/outlookcli.git", branch: "main", shallow: true

  # 发版后取消注释并填写 sha256（不要用占位符）：
  # url "https://github.com/ysu03zyy/outlookcli/archive/refs/tags/v0.1.0.tar.gz"
  # sha256 "................................"

  depends_on "go" => :build

  def install
    ldflags = "-s -w -X main.version=#{version}"
    system "go", "build", *std_go_args(ldflags: ldflags), "./cmd/outlookcli"
  end

  test do
    assert_match "Outlook mail", shell_output("#{bin}/outlookcli --help")
  end
end
