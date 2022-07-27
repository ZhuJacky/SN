FROM alpine

EXPOSE 8000


WORKDIR /opt/admin

COPY ./Shanghai /etc/localtime
CMD ["cd /opt/admin/"]
#CMD ["/opt/admin/go-admin","server","-c", "/opt/admin/config/settings.yml"]
CMD ["./go-admin","server","-c", "config/settings.yml"]
