FROM gcr.io/distroless/static

USER 1001

COPY dist/controller /usr/local/bin/controller

EXPOSE 8443 8443

ENTRYPOINT ["controller"]