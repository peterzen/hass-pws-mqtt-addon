# https://developers.home-assistant.io/docs/add-ons/configuration#add-on-dockerfile
ARG BUILD_FROM

FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.21 AS build
WORKDIR /root/
COPY ./src/ .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s"


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
