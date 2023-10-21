migrate create -ext sql -dir pkg/app/migrations -seq users
migrate create -ext sql -dir pkg/app/migrations -seq sessions
migrate create -ext sql -dir pkg/app/migrations -seq password_reset
migrate create -ext sql -dir pkg/app/migrations -seq galleries
migrate create -ext sql -dir pkg/app/migrations -seq galleries_publish

migrate -source file://pkg/app/migrations -database postgres://sa:"@dmin1234"@localhost:5432/lenslocked?sslmode=disable up
migrate -source file://pkg/app/migrations -database postgres://sa:"@dmin1234"@localhost:5432/lenslocked?sslmode=disable down
