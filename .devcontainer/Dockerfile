FROM fedora:42

SHELL ["/bin/bash", "-c"]

RUN \
  sed -e 's|^metalink=|#metalink=|g' \
  -e 's|^#baseurl=http://download.example/pub/fedora/linux|baseurl=https://mirrors.tuna.tsinghua.edu.cn/fedora|g' \
  -i.bak \
  /etc/yum.repos.d/fedora.repo \
  /etc/yum.repos.d/fedora-updates.repo

RUN dnf makecache && dnf install --setopt=install_weak_deps=False -y which unzip git bash-completion

ARG GO_VERSION=1.25.3
RUN \
  curl -sS https://webi.sh/golang@${GO_VERSION} | sh && source ~/.config/envman/PATH.env && \
  go env -w GOPROXY=https://goproxy.cn,direct

RUN \
  dnf copr enable -y atim/starship && dnf install --setopt=install_weak_deps=False -y starship && \
  echo -e "\n# starship\neval \"\$(starship init bash)\"\n" >> /root/.bashrc && \
  mkdir -p /root/.config && starship preset pure-preset -o /root/.config/starship.toml
