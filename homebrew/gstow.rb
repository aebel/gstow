class Gstow < Formula
  desc "Dotfile manager inspired by GNU Stow with per-package configuration"
  homepage "https://github.com/aebel/gstow"
  version "0.1.0"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/aebel/gstow/releases/download/v#{version}/gstow_#{version}_darwin_amd64.tar.gz"
      sha256 "b4208911316b3194f59e593cc16ede63822aabd41887981eeaef84f8f7160bad"
    end
    on_arm do
      url "https://github.com/aebel/gstow/releases/download/v#{version}/gstow_#{version}_darwin_arm64.tar.gz"
      sha256 "1041cc4c35a86ddc77759a5c054ed31ce479c44f5f40dc86fa9948ab1b96da5f"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/aebel/gstow/releases/download/v#{version}/gstow_#{version}_linux_amd64.tar.gz"
      sha256 "929d234e2058309f267b3b19506885e4b93f5d058b8cf6bade1ccb22b3f35f33"
    end
    on_arm do
      url "https://github.com/aebel/gstow/releases/download/v#{version}/gstow_#{version}_linux_arm64.tar.gz"
      sha256 "237cbb6b5c6bfff5aa257c855e1ef770f35f8b70ec371e863fbb1d9ec0304972"
    end
  end

  def install
    bin.install "gstow"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/gstow -V")
  end
end
