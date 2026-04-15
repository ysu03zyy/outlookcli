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
  version "0.1.1"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/ysu03zyy/outlookcli/releases/download/v#{version}/outlookcli_#{version}_darwin_arm64.tar.gz"
      sha256 "701a20f68a3c4c24f74ae400b55a4c82a818ddbb7db3715c9b785122d05d9f55"
    else
      url "https://github.com/ysu03zyy/outlookcli/releases/download/v#{version}/outlookcli_#{version}_darwin_amd64.tar.gz"
      sha256 "840bd5c484a3a80e683e0dd6f7bed9e192f16b7444fbe5040bf24804ecbd2b28"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/ysu03zyy/outlookcli/releases/download/v#{version}/outlookcli_#{version}_linux_arm64.tar.gz"
      sha256 "9c74afebb829deb689abc8eb6605e6415fbef662c1649206ff0cd74b8fce4184"
    else
      url "https://github.com/ysu03zyy/outlookcli/releases/download/v#{version}/outlookcli_#{version}_linux_amd64.tar.gz"
      sha256 "e891e4978d1ad99b1bb9d0a74f18e277d9f87a3841702d72d6ed7d212599dd90"
    end
  end

  def install
    bin.install "outlookcli"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/outlookcli --version")
  end
end
