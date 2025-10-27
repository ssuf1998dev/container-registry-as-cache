FROM fedora:42

SHELL ["/bin/bash", "-c"]

RUN dnf install --setopt=install_weak_deps=False -y which unzip git bash-completion

ARG GO_VERSION=1.25.3
RUN \
  curl -sS https://webi.sh/golang@${GO_VERSION} | sh && source ~/.config/envman/PATH.env && \
  go env -w GOPROXY=https://goproxy.cn,direct

RUN \
  curl -sS https://starship.rs/install.sh | sh -s -- --yes && \
  echo -e "\n# starship\neval \"\$(starship init bash)\"\n" >> /root/.bashrc && \
  mkdir -p /root/.config && starship preset pure-preset -o /root/.config/starship.toml
