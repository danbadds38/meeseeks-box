FROM arm32v6/alpine:3.7
 
RUN apk --no-cache add ca-certificates

COPY meeseeks-box /
 
ENTRYPOINT [ "/meeseeks-box" ]
