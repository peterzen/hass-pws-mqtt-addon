
# https://developers.home-assistant.io/docs/add-ons/configuration#add-on-dockerfile
ARG BUILD_FROM
ARG BUILD_ARCH

FROM golang AS build
ARG BUILD_ARCH
WORKDIR /root/
COPY ./src/ .
ENV BUILD_ARCH=${BUILD_ARCH}
RUN ./build.sh


FROM $BUILD_FROM

ARG TEMPIO_VERSION 
ARG BUILD_ARCH

# Execute during the build of the image
RUN \
    curl -sSLf -o /usr/bin/tempio \
    "https://github.com/home-assistant/tempio/releases/download/${TEMPIO_VERSION}/tempio_${BUILD_ARCH}"

# Copy root filesystem
COPY rootfs /

# Copy pwsmqttdispatcher binary
COPY --from=build /root/pwsmqttdispatcher /usr/bin

ENTRYPOINT [ "/usr/bin/pwsmqttdispatcher" ]