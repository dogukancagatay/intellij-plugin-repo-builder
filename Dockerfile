FROM docker.io/alpine:3.14
LABEL maintainer="Dogukan Cagatay <dcagatay@gmail.com>"

ARG APP_VERSION=v1.0.2

ENV SERVER_URL "http://localhost:3000"

# Get dependencies
RUN apk add --no-cache \
  bash \
  curl

WORKDIR /app

# Download intellij idea
#ARG INTELLIJ_VERSION=2022.2.1
#RUN curl -fSL --retry 3 "https://download.jetbrains.com/idea/ideaIU-${INTELLIJ_VERSION}.exe" \
#  -o /app/out/ideaIU-${INTELLIJ_VERSION}.exe

# Download and unarchive repo builder for serving
RUN curl -fSL --retry 3 "https://github.com/dogukancagatay/intellij-plugin-repo-builder/releases/download/${APP_VERSION}/intellij-plugin-repo-builder-${APP_VERSION}-linux_amd64.tar.gz" \
    -o /tmp/intellij-plugin-repo-builder-${APP_VERSION}-linux_amd64.tar.gz && \
  tar -xf /tmp/intellij-plugin-repo-builder-${APP_VERSION}-linux_amd64.tar.gz -C /tmp/ && \
  mv /tmp/intellij-plugin-repo-builder-${APP_VERSION}-linux_amd64/repo-builder /app/ && \
  rm -rf /tmp/intellij-plugin-repo-builder-${APP_VERSION}-linux_amd64*

# Build repo
COPY ./config.yaml ./
RUN ./repo-builder -build

COPY ./entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD ["./repo-builder", "-serve"]
