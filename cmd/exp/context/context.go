package main

import (
	stdctx "context"
	"fmt"

	"github.com/AguilaMike/lenslocked/pkg/app/context"
	"github.com/AguilaMike/lenslocked/pkg/app/models"
)

func main() {
	ctx := stdctx.Background()

	user := models.User{
		Email: "jon@calhoun.io",
	}

	ctx = context.WithUser(ctx, &user)
	retrievedUser := context.User(ctx)
	fmt.Println(retrievedUser.Email)
}
