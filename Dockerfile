FROM scratch

COPY ./main /main

EXPOSE 8000
CMD ["/main"]
