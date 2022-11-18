FROM golang AS build
WORKDIR /src
COPY *.go go.mod ./
RUN go mod tidy && go build -v -o /out/floating-eye .
FROM opensuse/tumbleweed AS prod
WORKDIR /bot
ARG TZ="Europe/Moscow"
ENV TZ=${TZ}
RUN zypper -qn install -y bsd-games ca-certificates xplanet
COPY xplanet ./xplanet
COPY monsters ./monsters
COPY --from=build /out/floating-eye .
CMD ["/bot/floating-eye"]
