package baseimage

const BaseDockerfile = `FROM ubuntu:22.04

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
RUN useradd -m -s /bin/bash $USER && \
    echo "$USER ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers && \
    mkdir -p /home/$USER/.ssh && \
    chmod 700 /home/$USER/.ssh && \
    chown -R $USER:$USER /home/$USER/.ssh && \
    mkdir -p /workspace && \
    chown -R $USER:$USER /workspace

# Configure SSH
RUN mkdir -p /var/run/sshd && \
    sed -i 's/#PermitRootLogin.*/PermitRootLogin no/' /etc/ssh/sshd_config && \
    sed -i 's/#PubkeyAuthentication.*/PubkeyAuthentication yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config && \
    echo "AllowUsers dev" >> /etc/ssh/sshd_config

# SSH agent forwarding support
RUN echo 'Host *\n    ForwardAgent yes' >> /etc/ssh/ssh_config

# Create entrypoint script
RUN echo '#!/bin/bash\nset -e\nif [ ! -f /etc/ssh/ssh_host_rsa_key ]; then\n    ssh-keygen -A\nfi\nexec "$@"' > /entrypoint.sh && \
    chmod +x /entrypoint.sh

WORKDIR /workspace

EXPOSE 22

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/usr/sbin/sshd", "-D"]
`

const ImageName = "devc-base"
const ImageTag = "latest"
const FullImageName = ImageName + ":" + ImageTag
