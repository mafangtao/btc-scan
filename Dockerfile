FROM golang:1.10.1 AS builder

ENV project_dir=/go/src/github.com/liyue201/btc-scan
WORKDIR /app

COPY . ${project_dir}
RUN go build -o  ${project_dir}/output/btc-scan ${project_dir}/main.go

FROM centos:7
ENV project_dir /go/src/github.com/liyue201/btc-scan
ENV work_dir  /server

WORKDIR ${work_dir}
COPY --from=builder ${project_dir}/output ${work_dir}/${app}

CMD ["/server/btc-scan", "-C", "configs/scan.yml"]
