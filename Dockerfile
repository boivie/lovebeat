FROM busybox
COPY lovebeat /bin/lovebeat

RUN mkdir /data && chown default:default /data

VOLUME ["/data", "/etc/lovebeat.conf.d"]
WORKDIR /data

EXPOSE 8127/udp 8127/tcp
EXPOSE 8080

USER default

ENTRYPOINT ["/bin/lovebeat"]
