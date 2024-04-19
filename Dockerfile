FROM golang:bookworm

WORKDIR /claude

COPY . /claude

CMD [ "./build.sh" ]