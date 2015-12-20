FROM alpine:3.2
ADD out/main /
cmd ["/main"]
EXPOSE 3000
