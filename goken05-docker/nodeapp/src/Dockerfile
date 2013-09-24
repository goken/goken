FROM    base
RUN     apt-get update
RUN     apt-get install -q -y nodejs

ADD     . /src
EXPOSE  8080
CMD     ["node", "/src/app.js"]

