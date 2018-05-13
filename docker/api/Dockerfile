FROM scratch

LABEL maintainer="John Weldon <johnweldon4@gmail.com>" \
      company="Tempus Breve Software" \
      description="OAuth2 API that emulates Gitlab"

COPY public /public/
COPY data /data/
ADD api api

ENV PORT 4884
ENV PUBLIC_HOME /public
ENV VERBOSE=false
ENV DATA_HOME /data
ENV CLIENTS /data/clients.json
ENV USERS /data/users.json

EXPOSE 4884

ENTRYPOINT ["/api", "serve"]
