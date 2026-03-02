class Gstow < Formula
  desc "Dotfile manager inspired by GNU Stow with per-package configuration"
  homepage "https://github.com/aebel/gstow"
  version "0.1.0"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/aebel/gstow/releases/download/v#{version}/gstow_#{version}_darwin_amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256"
    end
    on_arm do
      url "https://github.com/aebel/gstow/releases/download/v#{version}/gstow_#{version}_darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/aebel/gstow/releases/download/v#{version}/gstow_#{version}_linux_amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256"
    end
    on_arm do
      url "https://github.com/aebel/gstow/releases/download/v#{version}/gstow_#{version}_linux_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256"
    end
  end

  def install
    bin.install "gstow"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/gstow --version")
  end
end
