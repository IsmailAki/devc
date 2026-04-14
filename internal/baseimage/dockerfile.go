package baseimage

import (
	"encoding/base64"
	"fmt"
)

const entrypointScript = `#!/bin/bash
set -euo pipefail

align_dev_user() {
    local target_uid="${DEVC_UID:-}"
    local target_gid="${DEVC_GID:-}"

    if [ -z "$target_uid" ] || [ -z "$target_gid" ]; then
        return
    fi

    if ! [[ "$target_uid" =~ ^[0-9]+$ ]] || ! [[ "$target_gid" =~ ^[0-9]+$ ]]; then
        return
    fi

    local current_uid current_gid
    current_uid="$(id -u dev)"
    current_gid="$(id -g dev)"

    if [ "$current_gid" != "$target_gid" ]; then
        if getent group "$target_gid" >/dev/null 2>&1; then
            usermod -g "$target_gid" dev
        else
            groupmod -g "$target_gid" dev
            usermod -g "$target_gid" dev
        fi
    fi

    if [ "$current_uid" != "$target_uid" ]; then
        usermod -u "$target_uid" dev
    fi

    chown -R dev:$(id -gn dev) /home/dev
}

prepare_workspace() {
    mkdir -p /workspace /var/run/sshd
    if [ "$(stat -c '%u:%g' /workspace)" = "0:0" ]; then
        chown dev:$(id -gn dev) /workspace
    fi
}

start_dind() {
    if [ "${DEVC_ENABLE_DIND:-0}" != "1" ]; then
        return
    fi

    mkdir -p /var/lib/docker /var/log

    if ! getent group docker >/dev/null 2>&1; then
        groupadd docker
    fi

    usermod -aG docker dev

    if ! pgrep dockerd >/dev/null 2>&1; then
        dockerd \
            --host=unix:///var/run/docker.sock \
            --group=docker \
            --data-root=/var/lib/docker \
            >/var/log/dockerd.log 2>&1 &
    fi

    local attempts=0
    until docker info >/dev/null 2>&1; do
        attempts=$((attempts + 1))
        if [ "$attempts" -ge 60 ]; then
            cat /var/log/dockerd.log >&2 || true
            echo "dockerd failed to start" >&2
            exit 1
        fi
        sleep 1
    done
}

if [ ! -f /etc/ssh/ssh_host_rsa_key ]; then
    ssh-keygen -A
fi

align_dev_user
prepare_workspace
start_dind

exec "$@"
`

var BaseDockerfile = fmt.Sprintf(`FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive
ENV USER=dev

RUN apt-get update && apt-get install -y \
    git \
    curl \
    wget \
    build-essential \
    sudo \
    openssh-server \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user with sudo access
RUN groupadd --gid 1000 $USER && \
    useradd --uid 1000 --gid 1000 -m -s /bin/bash $USER && \
    echo "$USER ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers && \
    mkdir -p /home/$USER/.ssh && \
    chmod 700 /home/$USER/.ssh && \
    chown -R $USER:$USER /home/$USER/.ssh && \
    mkdir -p /workspace && \
    chown -R $USER:$USER /workspace

# Configure SSH
RUN mkdir -p /var/run/sshd && \
    sed -i 's/#PermitRootLogin.*/PermitRootLogin prohibit-password/' /etc/ssh/sshd_config && \
    sed -i 's/#PubkeyAuthentication.*/PubkeyAuthentication yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config && \
    echo "AllowUsers dev root" >> /etc/ssh/sshd_config

# SSH agent forwarding support
RUN echo 'Host *\n    ForwardAgent yes' >> /etc/ssh/ssh_config

# Create entrypoint script
RUN printf '%%s' '%s' | base64 -d > /usr/local/bin/devc-entrypoint.sh && \
    chmod +x /usr/local/bin/devc-entrypoint.sh

WORKDIR /workspace

EXPOSE 22

ENTRYPOINT ["/usr/local/bin/devc-entrypoint.sh"]
CMD ["/usr/sbin/sshd", "-D"]
`, base64.StdEncoding.EncodeToString([]byte(entrypointScript)))

const ImageName = "devc-base"
const ImageTag = "v3"
const FullImageName = ImageName + ":" + ImageTag
