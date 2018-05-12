FROM opensuse:42.3
MAINTAINER Sergey Koyushev <serjk91@gmail.com>
ARG GO_URL=https://storage.googleapis.com/golang/go1.9.2.linux-amd64.tar.gz

RUN zypper --quiet --non-interactive in curl make git gcc libpacemaker-devel libxml2-devel glib2-devel pacemaker

RUN curl -sSL ${GO_URL} > /tmp/golang.tar.gz  \
 && tar -C /usr/local/ -xf /tmp/golang.tar.gz \
 && rm /tmp/golang.tar.gz                     \
 && mkdir -p /opt/go/src /opt/go/bin          \
 && chmod -R 777 /opt/go

ENV PATH="$PATH:/sbin:/usr/sbin:/bin:/usr/bin:/opt/go/bin:/usr/local/go/bin" GOROOT=/usr/local/go GOPATH=/opt/go DEVROOT=/root DEVPATH=/root

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

WORKDIR $DEVPATH