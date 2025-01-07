FROM eclipse-temurin:11-jre-alpine

RUN apk add --no-cache fontconfig ttf-dejavu
RUN ln -s /usr/lib/libfontconfig.so.1 /usr/lib/libfontconfig.so && \
    ln -s /lib/libuuid.so.1 /usr/lib/libuuid.so.1 && \
    ln -s /lib/libc.musl-x86_64.so.1 /usr/lib/libc.musl-x86_64.so.1
ENV LD_LIBRARY_PATH /usr/lib
ENV APPLICATION_USER=wca

RUN adduser -D -g '' $APPLICATION_USER
RUN mkdir /app
RUN chown -R $APPLICATION_USER /app


USER $APPLICATION_USER

WORKDIR /app
ARG ONLINE
ENV ONLINE_MODE=${ONLINE:+"--online"}
ENV ONLINE_MODE=${ONLINE_MODE:-""}
ADD TNoodle-WCA-1.2.2.jar /app/TNoodle-WCA.jar

EXPOSE 2014
ENV PORT=${PORT:-2014}


CMD java -server -XX:+UnlockExperimentalVMOptions -XX:InitialRAMFraction=2 -XX:MinRAMFraction=2 -XX:MaxRAMFraction=2 -XX:+UseG1GC -XX:MaxGCPauseMillis=100 -XX:+UseStringDeduplication -jar TNoodle-WCA.jar $ONLINE_MODE -b


#CMD ["java", "-jar", "TNoodle-WCA.jar"]
#CMD ["sh", "-c", "Xvfb :99 -screen 0 1024x768x16 & java -Djava.awt.headless=false -jar TNoodle-WCA.jar"]

# 需要用java sdk 11

# build: docker build -t tnoodle:1.2.2 .
# run: docker run -d --restart always -p 2014:2014 --name tnoodle tnoodle:1.2.2



