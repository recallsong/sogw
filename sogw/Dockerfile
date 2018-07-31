FROM alpine:3.6

RUN mkdir -p /sogw/conf /sogw/logs
Copy bin/linux-amd64-sogw /sogw/sogw
Copy conf/meta.example.json /sogw/conf/meta.json
Copy conf/sogw.example.yml /sogw/conf/sogw.yml

WORKDIR /sogw
VOLUME /sogw/conf

CMD ["/sogw/sogw"]