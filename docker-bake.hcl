variable "IMAGE_NAME" {
    default = "dcagatay/intellij-plugin-repo-builder"
}

variable "APP_VERSION" {
    default = "v1.0.2"
}

variable "INTELLIJ_VERSION" {
    default = "2022.2.1"
}

group "default" {
    targets = [ "latest" ]
}

target "latest" {
    context = "."
    platforms = [ "linux/amd64" ]
    dockerfile = "Dockerfile"
    args = {
        APP_VERSION = APP_VERSION
        INTELLIJ_VERSION = INTELLIJ_VERSION
    }
    tags = [
        "docker.io/${IMAGE_NAME}:latest",
        "docker.io/${IMAGE_NAME}:${APP_VERSION}-${INTELLIJ_VERSION}"
    ]
}
