ARG PLAYWRIGHT_VERSION=v1.52.0-noble
ARG PLAYWRIGHT_PACKAGE=1.52.0
FROM mcr.microsoft.com/playwright:${PLAYWRIGHT_VERSION}

USER root
RUN apt-get update \
    && apt-get install -y --no-install-recommends x11vnc x11-utils \
    && rm -rf /var/lib/apt/lists/*
USER pwuser
WORKDIR /home/pwuser
COPY launch-headed-browser.js /home/pwuser/launch-headed-browser.js
ARG PLAYWRIGHT_PACKAGE
RUN npm init -y \
    && npm install --save-exact "playwright@${PLAYWRIGHT_PACKAGE}" \
    && node node_modules/playwright/cli.js --version
