FROM nginx:1.25
# RUN mkdir /etc/nginx/cert
# COPY cert/server-cert.pem /etc/nginx/cert
# COPY cert/server-key.pem /etc/nginx/cert
# COPY cert/ca-cert.pem /etc/nginx/cert
COPY nginx.conf /etc/nginx/nginx.conf

CMD ["nginx", "-g", "daemon off;"]