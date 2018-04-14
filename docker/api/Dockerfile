FROM scratch

LABEL maintainer="John Weldon <johnweldon4@gmail.com>" \
      company="Tempus Breve Software" \
      description="OAuth2 API that emulates Gitlab"

COPY public /public/
ADD api api

ENV PORT 4884
ENV PUBLIC_DIR /public
ENV VERBOSE=false

EXPOSE 4884

ENTRYPOINT ["/api"]
