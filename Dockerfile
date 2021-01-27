# iron/go is the alpine image with only ca-certificates added
FROM photon:3.0
WORKDIR /app
# add static html files
ADD static /app/static
# Now just add the binary
ADD tanzu-observability-collector-for-aws-fargate_linux /app/
ENTRYPOINT ["./tanzu-observability-collector-for-aws-fargate_linux"]
