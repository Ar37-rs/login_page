# Login page + google oauth2 

Clone repo:

```bash
git clone https://github.com/Ar37-rs/login_page_plus_oauth2.git
cd login_page_plus_oauth2
```

A note, before running:
Change the empty client_id and client_secret strings inside main.go file according to your GA ClientID and ClientSecret,
and also you have to add the url: http://localhost:1323/auth/oauth2 as one of your Authorised redirect URIs on your GA credential config.
for more info visit: https://console.cloud.google.com/apis/credentials

How to run (the SPA and database/sqlite3 are already included and configured):

```bash
cd backend
go mod tidy
go run main.go
```

and then open http://localhost:1323

![alt_test](image/flowchart.svg)
