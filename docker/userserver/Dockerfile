FROM scratch

LABEL company="Tempus Breve Software" \
      maintainer="John Weldon <johnweldon4@gmail.com>" \
      description="User API that emulates Gitlab"

ADD userserver userserver

ENV PORT 4884
ENV VERBOSE=false

EXPOSE 4884

ENTRYPOINT ["/userserver"]
