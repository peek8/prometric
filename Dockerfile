
# Get the minimum base
FROM alpine:3.22 

# This binary is coming from goreleaser
# GoReleaser builds the binary first and then injects it into the Docker build context when it runs docker build.
COPY prometric /usr/local/bin/prometric


# Provide a non-root user (distroless provides user 65532)
USER 65532:65532


ENTRYPOINT ["/usr/local/bin/prometric"]