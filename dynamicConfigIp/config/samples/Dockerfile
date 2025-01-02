# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM ubuntu:1.1
ADD bin/manager  /usr/local/bin
USER 65532:65532

ENTRYPOINT ["usr/local/bin/manager"]
