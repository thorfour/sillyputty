FROM ubuntu
EXPOSE 80 443
COPY ca-certificates.crt /etc/ssl/certs/
COPY server /
CMD ["/server"]
