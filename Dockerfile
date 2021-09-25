FROM docker.io/alpine:3.14
LABEL maintainer="Dogukan Cagatay <dcagatay@gmail.com>"

ARG APP_VERSION=v1.0.1
ARG INTELLIJ_VERSION=2021.2.2

ENV SERVER_URL "http://localhost:3000"

# Get dependencies
RUN apk add --no-cache \
  bash \
  curl

WORKDIR /app/out

# Download intellij idea
RUN curl -fSL --retry 3 "https://download.jetbrains.com/idea/ideaIU-${INTELLIJ_VERSION}.exe" \
  -o ideaIU-${INTELLIJ_VERSION}.exe

# Download repo builders
RUN curl -fSL --retry 3 "https://github.com/dogukancagatay/intellij-plugin-repo-builder/releases/download/${APP_VERSION}/intellij-plugin-repo-builder-${APP_VERSION}-linux_amd64.tar.gz" \
  -o intellij-plugin-repo-builder-${APP_VERSION}-linux_amd64.tar.gz && \
  curl -fSL --retry 3 "https://github.com/dogukancagatay/intellij-plugin-repo-builder/releases/download/${APP_VERSION}/intellij-plugin-repo-builder-${APP_VERSION}-windows_amd64.tar.gz" \
  -o intellij-plugin-repo-builder-${APP_VERSION}-windows_amd64.tar.gz

# Unarchive repo builder for serving
RUN tar -xf intellij-plugin-repo-builder-${APP_VERSION}-linux_amd64.tar.gz -C /tmp/ && \
  mv /tmp/intellij-plugin-repo-builder*/* /app/ && \
  rm -rf /tmp/intellij-plugin-repo-builder*

# Create index.html for serving files
RUN echo '<html><body>' >> index.html && \
  ls -1 | grep -v index.html | xargs -Ixx echo '<p><a href="http://localhost:3000/xx">xx</a></p>' >> index.html && \
  echo '</body</html>' >> index.html

WORKDIR /app

# # Build repo
RUN cd /app && \
  ls -al && \
  ./repo-builder -build

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD ["./repo-builder", "-serve"]
