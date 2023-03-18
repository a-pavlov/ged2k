FROM ubuntu:16.04

RUN apt-get update -y && apt-get upgrade -y && \
    apt-get install -y software-properties-common \
                       locales \
                       apt-utils \
                       git \
                       build-essential \
                       bash \
                       wget \
                       curl \
                       unzip libzstd-dev zstd \
                       g++ \
                       python-dev \
                       autotools-dev \
                       libicu-dev \
                       libbz2-dev \
                       autoconf \
                       libtool \
                       libpthread-stubs0-dev \
                       libpcre3-dev \
                       openssl \
                       autogen \
                       libssl-dev \
                       less \
                       glibc-source

RUN apt-get install -y  binutils-dev \
                        libcrypto++-dev \
                        libcrypto++-utils \
                        libgtk2.0-dev \
                        libgd2-xpm-dev \
                        libgeoip-dev \
                        zlib1g-dev \
                        autopoint

RUN apt-get install -y gtk2-engines

#RUN cd /tmp && \
#    wget http://prdownloads.sourceforge.net/wxwindows/wxWidgets-2.8.12.tar.gz &&\
#    tar -xzf wxWidgets-2.8.12.tar.gz &&\
#    cd wxWidgets-2.8.12 &&\
#    ./configure --with-gtk --without-libtiff --without-subdirs --enable-unicode --enable-optimise --enable-xrc --with-x &&\
#    make && make install

RUN cd /tmp &&\
    wget https://github.com/wxWidgets/wxWidgets/releases/download/v3.0.5/wxWidgets-3.0.5.tar.bz2 &&\
    tar -xf wxWidgets-3.0.5.tar.bz2 &&\
    cd wxWidgets-3.0.5 &&\
    ./configure --prefix=/usr --with-gtk --enable-unicode --disable-compat24 --enable-optimise &&\
    make && make install &&\
    cd /

RUN ldconfig

RUN ( \
    echo 'LogLevel DEBUG2'; \
    echo 'PermitRootLogin yes'; \
    echo 'PasswordAuthentication yes'; \
    echo 'Subsystem sftp /usr/lib/openssh/sftp-server'; \
  ) > /etc/ssh/sshd_config_user \
  && mkdir /run/sshd

RUN useradd -m user \
  && yes pwd | passwd user

RUN usermod -s /bin/bash user

RUN apt-get install -y ssh

CMD ["/usr/sbin/sshd", "-D", "-e", "-f", "/etc/ssh/sshd_config_user"]

#
# wget https://github.com/amule-project/amule/archive/refs/tags/2.3.3.tar.gz
# tar -xzf amule-2.3.3.tar.gz
# cd aMule-2.3.2
# ./configure --enable-debug --enable-optimize




