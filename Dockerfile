FROM scratch
ADD release/passwd.minimal /etc/passwd
ADD lovebeat /lovebeat
USER nobody

VOLUME /data
WORKDIR /data

EXPOSE 8127/udp 8127/tcp
EXPOSE 8080

CMD ["/lovebeat"]
