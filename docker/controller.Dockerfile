FROM gcr.io/distroless/static@sha256:d6fa9db9548b5772860fecddb11d84f9ebd7e0321c0cb3c02870402680cc315f

USER 1001

COPY dist/controller /usr/local/bin/controller

EXPOSE 8443 8443

ENTRYPOINT ["controller"]