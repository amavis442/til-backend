## Today I learned app

Small and simple project for keeping track of the things i learned today and a todo list. You learn every day, but my problem is, i also forget a lot.

I use docker for development (postgresql) and not production, so it can be a bit tricky.
Goal is:

- [x] Add new TIL
- [x] Update TIL
- [x] User login and roles
- [x] Search for TILL based on category and title
- [ ] Category view 
- [ ] Export TIL to markdown file
- [x] Pagination
- [ ] Todo keep track of what needs to be done
- [x] KISS
- [x] DRY
- [x] Clean Architecture
- [x] Fun

For the backend i use **go**. Frontend is build in **Sveltekit**.

For database migration i use **go-migrate**, so you will need to download this. Fo easy compiling and migrating i use **make** and **Makefile**.

Maybe i will make a flutter app that can consume the api. Will see.


# Installation

For this you need [golang](https://go.dev/dl/go1.24.5.linux-amd64.tar.gz) version 1.22 or higher.

## Build 

> git clone https://github.com/amavis442/til-backend.git

> go build -o cmd/server/server cmd/server/main.go

This will install all the needed packages.

## Generate keys

Do not forget to generate the jwt tokens with openssl

> chmod +x generate-jwt.sh

> ./generate-jwt.sh

This will generate the tokens **private.pem** and **public.pem** and place them in config/jwt.

## Config

Then you will have to copy **.env** to **.env.local** and change **DB_DSN** to your postgresql server.

>DB_DSN=host=db user=tiluser password=tilpassword dbname=til port=5432 sslmode=disable

If you want the server to run on another port, change `PORT=3031` to the desired port.

To start the app `cmd/server/server` but not before you followed the steps below.

## Migrate database
### Create database
To create a database in linux (debian/ubuntu)

> sudo postgres

> psql

> CREATE ROLE rolename/username WITH LOGIN CREATEDB PASSWORD 'password';

> CREATE DATABASE til WITH OWNER 'rolename/username';

For UTF8 encoding if not already

> CREATE DATABASE dbname WITH ENCODING 'UTF8' LOCALE='C.UTF-8' TEMPLATE=template0 OWNER rolename/username;

### Create tables

In folder migrations you will find the migration files to create the tables for the database.
If you use [golang-migrate](https://pkg.go.dev/github.com/golang-migrate/migrate/v4), you can use something like

> migrate -path migrations -database "postgres://dbuser:dbpasswd@db:5432/dbname?sslmode=disable" up

Replace *dbuser*,*dbpasswd* and *dbname* with your setup. This will update your database.

These steps can also be done in ui on windows with [pgAdmin 4](https://www.pgadmin.org/download/pgadmin-4-windows/).

## Endpoints of api

This is the backend only and has an api which you can find in cm/server/main.go:

### Unprotected:

```
http://localhost:3031/auth/register (POST) // username, email and password

http://localhost:3031/auth/login (POST)

http://localhost:3031/auth/refresh-token (POST) // get a new access and refresh token
```

### Protected (needs access token):

```
http://localhost:3031/api/tils (GET) // get a list of all til entries

http://localhost:3031/api/tils/search (POST) // search title and/or category

http://localhost:3031/api/tils/:id (GET) // get a til entry

http://localhost:3031/api/tils/ (POST) create a til entry

http://localhost:3031/api/tils/:id (PUT) // update til entry
```

## Start the api server

Linux terminal

> CORS_ALLOWED_ORIGIN=http://localhost cmd/server/server

or

> CORS_ALLOWED_ORIGIN=http://localhost go run cmd/server/main.go

In powershell

> ENV CORS_ALLOWED_ORIGIN=http://localhost cmd/server/server

or 

> ENV CORS_ALLOWED_ORIGIN=http://localhost go run cmd/server/main.go

If you run a server with ip `192.168.2.60` then change `http://localhost` to `http://192.168.2.60` or `https://` reps.