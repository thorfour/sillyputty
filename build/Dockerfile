FROM scratch
EXPOSE 443
EXPOSE 80
COPY ca-certificates.crt /etc/ssl/certs/
COPY server /
CMD ["/server"]
