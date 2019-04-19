FROM gcr.io/moonrhythm-containers/alpine:3.9

WORKDIR /app

COPY pikkanode ./
EXPOSE 8080

ENTRYPOINT ["/app/pikkanode"]
