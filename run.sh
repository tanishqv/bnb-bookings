#!/bin/bash

go build -o bookings cmd/web/*.go
./bookings -dbname=bookings -dbuser=postgres -dbpwd=postgres -cache=false -production=false