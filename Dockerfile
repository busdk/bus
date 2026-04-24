# Standard test image for BusDK modules.
# Source code is mounted at runtime so local replace directives and sibling modules work.
FROM --platform=linux/amd64 golang:1.22-bookworm

RUN apt-get update && apt-get install -y --no-install-recommends \
    make \
    nodejs \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace
CMD ["make", "test"]
