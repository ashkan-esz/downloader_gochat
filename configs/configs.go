package configs

const (
	//todo : use env for db url
	DbUrl            string = "postgres://root:mysecretpassword@localhost:5432/go-chat?sslmode=disable"
	SigningSecretKey string = "secret"
)
