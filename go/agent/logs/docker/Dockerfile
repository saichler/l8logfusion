FROM saichler/builder:latest AS build

COPY main.go /home/src/github.com/saichler/build/main.go
RUN go mod init
RUN GOPROXY=direct GOPRIVATE=github.com go mod tidy
RUN go build -o logagent

FROM saichler/probler-security:latest AS final
COPY --from=build /home/src/github.com/saichler/build/logagent /home/run/logagent

ARG VNETIP='127.0.0.1'
ARG PATH='/data/logs'
ARG FILE='*'

ENTRYPOINT ["/home/run/logagent","$VNETIP","$PATH","$FILE"]