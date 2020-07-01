# iron/go is the alpine image with only ca-certificates added
FROM busybox:1.32-glibc
WORKDIR /app
# add static html files
ADD static /app/static
# Now just add the binary
ADD fargate_collector_linux /app/
ENTRYPOINT ["./fargate_collector_linux"]
