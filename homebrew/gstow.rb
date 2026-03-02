class Gstow < Formula
  desc "Dotfile manager inspired by GNU Stow with per-package configuration"
  homepage "https://github.com/aebel/gstow"
  version "0.1.1"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/aebel/gstow/releases/download/v#{version}/gstow_#{version}_darwin_amd64.tar.gz"
      sha256 "80adbb443694263921eee98621a9bb7861441745c751110c21c1ba3beeb4808c"
    end
    on_arm do
      url "https://github.com/aebel/gstow/releases/download/v#{version}/gstow_#{version}_darwin_arm64.tar.gz"
      sha256 "5f6c1307190cd0d760c3539b8ee0c40f2ef78c34eaec298439e2b4b180ebb520"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/aebel/gstow/releases/download/v#{version}/gstow_#{version}_linux_amd64.tar.gz"
      sha256 "fe04236df321b158843fa64369a527e3078234eecb3b49c7f874da0bcbd8b133"
    end
    on_arm do
      url "https://github.com/aebel/gstow/releases/download/v#{version}/gstow_#{version}_linux_arm64.tar.gz"
      sha256 "6eeafdd15df08911296f157dd154eb27aca50e107470fc9ccc295ecc53cf57cb"
    end
  end

  def install
    bin.install "gstow"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/gstow -V")
  end
end