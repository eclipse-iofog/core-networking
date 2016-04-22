FROM jeanblanchard/java:jdk-8u77

RUN mkdir /jar-file
COPY . /jar-file

CMD ["java", "-jar", "/jar-file/core-networking.jar"]
