go build -o bookings.exe cmd/web/.
bookings.exe -dbname=bookings -dbuser=postgres -dbpwd=postgres -cache=false -production=false