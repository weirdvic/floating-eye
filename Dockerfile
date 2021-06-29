FROM golang AS build
WORKDIR /src
COPY *.go go.mod config.json ./
RUN go mod tidy && go build -v -o /out/floating_eye .
FROM ubuntu AS prod
WORKDIR /bot
COPY monsters ./monsters
COPY xplanet ./xplanet
RUN echo "Europe/Moscow" > /etc/timezone && \
    apt update && \
    apt install -y bsdgames ca-certificates xplanet && \
    ln -fs /usr/games/pom /usr/bin/pom
COPY --from=build /out/floating_eye .
CMD ["/bot/floating_eye"]
