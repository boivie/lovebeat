FROM debian:jessie
RUN apt-get update && apt-get install --no-install-recommends -y ca-certificates && rm -rf /var/lib/apt/lists/* && apt-get clean

ADD https://github.com/boivie/lovebeat/releases/download/1.0.3/linux_amd64_lovebeat /bin/lovebeat
RUN chmod a+x /bin/lovebeat && mkdir /data

VOLUME ["/data", "/etc/lovebeat.conf.d"]
WORKDIR /data

EXPOSE 8127/udp 8127/tcp
EXPOSE 8080

ENTRYPOINT ["/bin/lovebeat"]
