FROM rockylinux:9.3
ARG USER_UID=10001
USER ${USER_UID}
COPY --chmod=755 ./client /app/client
ENTRYPOINT ["/app/client", "--host=grpc-server"]
