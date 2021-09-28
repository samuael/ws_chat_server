server {
        server_name chat.wsecho.golangapp.com;

        location / {

                proxy_pass http://localhost:8080;
        }
}
