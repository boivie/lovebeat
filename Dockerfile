FROM debian:jessie

ADD https://github.com/boivie/lovebeat/releases/download/1.0.1/linux_amd64_lovebeat /bin/lovebeat
RUN chmod a+x /bin/lovebeat && mkdir /data

VOLUME ["/data", "/etc/lovebeat.conf.d"]
WORKDIR /data

EXPOSE 8127/udp 8127/tcp
EXPOSE 8080

ENTRYPOINT ["/bin/lovebeat"]
