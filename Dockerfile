FROM scratch
#ADD ca-certificates.crt /etc/ssl/certs/
ADD out/main /
ADD tmp /tmp
ADD mime.types /etc/mime.types
cmd ["/main"]
EXPOSE 3000
