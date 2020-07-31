FROM alpine:latest
RUN echo "http://mirrors.tuna.tsinghua.edu.cn/alpine/edge/main" > /etc/apk/repositories && \
    echo "http://mirrors.tuna.tsinghua.edu.cn/alpine/edge/community" >> /etc/apk/repositories && \
    apk update && apk upgrade && \
    apk add --no-cache tzdata && \
    apk add --no-cache openldap-clients && \
    echo "Asia/shanghai" > /etc/timezone && \
    ln -s /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "hosts: files dns" > /etc/nsswitch.conf

ADD bin/ /opt/fusion/bin

EXPOSE 9080 9081
ENV BRANCH=dev

ENTRYPOINT ["/opt/fusion/bin/skel_main"]
