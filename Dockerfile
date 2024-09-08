FROM golang:1.21.1 as Builder

RUN mkdir /app
ADD . /app
WORKDIR /app
## We want to build our application's binary executable
RUN go mod download
RUN go build -o verve main.go


FROM golang:1.21.1
COPY --from=builder /app/verve .
RUN mkdir configs
ENV ENV=${ENV}
#COPY --from=builder /app/configs/* configs/
ENTRYPOINT ./verve