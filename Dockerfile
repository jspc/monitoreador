FROM alpine

COPY ./monitoreador /monitoreador
COPY ./entrypoint /entrypoint

EXPOSE 8000
CMD ["/entrypoint"]
