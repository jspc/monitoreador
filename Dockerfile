FROM alpine

COPY ./main /main
COPY ./entrypoint /entrypoint

EXPOSE 8000
CMD ["/entrypoint"]
