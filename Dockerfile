FROM golang:latest

# update os
RUN apt-get update && apt-get upgrade --yes

# install delve for vscode live debugging
RUN go get github.com/go-delve/delve/cmd/dlv

# install graphql lib
RUN go get github.com/machinebox/graphql

# forklift code to working dir
WORKDIR /app

COPY main.go ./

EXPOSE 8000 8001

CMD [ "go", "run", "main.go" ]