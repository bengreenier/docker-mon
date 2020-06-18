FROM nginx:alpine
RUN apk add curl
COPY ./nginx.conf /etc/nginx/nginx.conf
COPY ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
HEALTHCHECK --interval=10s --timeout=3s CMD curl -f http://localhost/ || exit 1
ENTRYPOINT ["/entrypoint.sh"]
