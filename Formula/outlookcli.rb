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
  version "0.1.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/ysu03zyy/outlookcli/releases/download/v#{version}/outlookcli_#{version}_darwin_arm64.tar.gz"
      sha256 "REPLACE_WITH_DARWIN_ARM64_SHA256"
    else
      url "https://github.com/ysu03zyy/outlookcli/releases/download/v#{version}/outlookcli_#{version}_darwin_amd64.tar.gz"
      sha256 "REPLACE_WITH_DARWIN_AMD64_SHA256"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/ysu03zyy/outlookcli/releases/download/v#{version}/outlookcli_#{version}_linux_arm64.tar.gz"
      sha256 "REPLACE_WITH_LINUX_ARM64_SHA256"
    else
      url "https://github.com/ysu03zyy/outlookcli/releases/download/v#{version}/outlookcli_#{version}_linux_amd64.tar.gz"
      sha256 "REPLACE_WITH_LINUX_AMD64_SHA256"
    end
  end

  def install
    bin.install "outlookcli"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/outlookcli --version")
  end
end
