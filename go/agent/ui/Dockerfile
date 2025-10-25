FROM saichler/builder:latest AS build

COPY main.go /home/src/github.com/saichler/build/main.go
RUN go mod init
RUN GOPROXY=direct GOPRIVATE=github.com go mod tidy
RUN go build -o logsui

FROM saichler/probler-security:latest AS final
COPY --from=build /home/src/github.com/saichler/build/logsui /home/run/logsui
COPY ./web /home/run/web

ENTRYPOINT ["/home/run/logsui"]