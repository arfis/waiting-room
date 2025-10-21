
FROM nghis.azurecr.io/nghis/tools/distroless
ARG SHORT_NAME

WORKDIR /app
COPY "/$SHORT_NAME" ./application
COPY "/open-api.yaml" ./open-api.yaml

ENTRYPOINT ["./application"]
