FROM openjdk:18-ea-11-jdk-alpine

ADD TNoodle-WCA-1.2.2.jar TNoodle-WCA.jar

EXPOSE 2014
CMD ["java", "-jar", "TNoodle-WCA.jar"]
# todo 需要用java sdk 11

# build: docker build -t tnoodle:1.2.2 .
# run: docker run -d --restart always -p 2014:2014 --name tnoodle tnoodle:1.2.2



