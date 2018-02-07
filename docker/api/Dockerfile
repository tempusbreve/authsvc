FROM scratch

MAINTAINER John Weldon <johnweldon4@gmail.com>

COPY public /public/
ADD api api

ENV PORT 4884
ENV PUBLIC_DIR /public
ENV VERBOSE=false

EXPOSE 4884

ENTRYPOINT ["/api"]
