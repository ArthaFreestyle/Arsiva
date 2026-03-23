start :
	go run cmd/web/main.go

migrate-fresh:
	psql "postgres://artha:passwordku@localhost:5432/arsiva?sslmode=disable" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	migrate -database "postgres://artha:passwordku@localhost:5432/arsiva?sslmode=disable" -path /home/artha/Documents/Arsiva/db/migrations_postgre down
	migrate -database "postgres://artha:passwordku@localhost:5432/arsiva?sslmode=disable" -path /home/artha/Documents/Arsiva/db/migrations_postgre up

migrate : 
	migrate -database "postgres://artha:passwordku@localhost:5432/arsiva?sslmode=disable" -path /home/artha/Documents/Arsiva/db/migrations_postgre up

migrate-down : 
	migrate -database "postgres://artha:passwordku@localhost:5432/arsiva?sslmode=disable" -path /home/artha/Documents/Arsiva/db/migrations_postgre down

seed : 
	migrate -database "postgres://artha:passwordku@localhost:5432/arsiva?sslmode=disable&x-migrations-table=schema_seeds" -path /home/artha/Documents/Arsiva/db/migrations_seed up

seed-down : 
	migrate -database "postgres://artha:passwordku@localhost:5432/arsiva?sslmode=disable&x-migrations-table=schema_seeds" -path /home/artha/Documents/Arsiva/db/migrations_seed down


