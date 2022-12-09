package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"entgo.io/bug/ent"
	_ "entgo.io/bug/ent/runtime"
	"entgo.io/bug/ent/schema"
	"entgo.io/bug/ent/user"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&_fk=1", ent.Debug(), ent.Log(func(s ...any) {
		fmt.Println(s...)
	}))
	if err != nil {
		panic(err)
	}
	defer client.Close()
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	ctx := context.Background()

	p1, err := client.Pet.Create().SetName("p1").Save(ctx)
	if err != nil {
		panic(err)
	}
	p2, err := client.Pet.Create().SetName("p2").Save(ctx)
	if err != nil {
		panic(err)
	}
	p3, err := client.Pet.Create().SetName("p3").Save(ctx)
	if err != nil {
		panic(err)
	}
	p4, err := client.Pet.Create().SetName("p4").Save(ctx)
	if err != nil {
		panic(err)
	}
	u1, err := client.User.Create().SetName("u1").SetAge(1).AddPets(p1, p2).Save(ctx)
	if err != nil {
		panic(err)
	}
	u2, err := client.User.Create().SetName("u2").SetAge(1).AddPets(p3, p4).Save(ctx)
	if err != nil {
		panic(err)
	}

	g, err := client.Group.Create().SetName("group").AddUsers(u1, u2).Save(ctx)
	if err != nil {
		panic(err)
	}

	{
		u1, err = client.User.Query().Where(user.Name("u1")).First(ctx)
		if err != nil {
			panic(err)
		}
		if err := client.User.DeleteOne(u1).Exec(ctx); err != nil {
			panic(err)
		}

		_, err = client.User.Query().Where(user.Name("u1")).First(ctx)
		if err == nil {
			panic("found no soft delete user")
		} else {
			if !ent.IsNotFound(err) {
				panic(err)
			}
		}

	}
	{
		if err := client.Pet.DeleteOne(p1).Exec(ctx); err != nil {
			panic(err)
		}
		if err := client.Pet.DeleteOne(p3).Exec(ctx); err != nil {
			panic(err)
		}

		{
			// normal query
			g, err = client.Group.Query().WithUsers(func(uq *ent.UserQuery) {
				uq.WithPets()
			}).First(ctx)
			if err != nil {
				panic(err)
			}
			data, err := json.MarshalIndent(g, "", "  ")
			if err != nil {
				panic(err)
			}
			fmt.Println(string(data))
		}
		{
			// query softdelete
			g, err = client.Group.Query().WithUsers(func(uq *ent.UserQuery) {
				uq.WithPets()
			}).First(schema.SkipSoftDelete(ctx))
			if err != nil {
				panic(err)
			}
			data, err := json.MarshalIndent(g, "", "  ")
			if err != nil {
				panic(err)
			}
			fmt.Println(string(data))
		}

	}
	{
		softdeletectx := schema.SkipSoftDelete(ctx)
		nu, err := client.User.Query().Where(user.Name("u1")).First(softdeletectx)
		if err != nil {
			panic(err)
		}
		fmt.Println(nu)
		if err := client.User.DeleteOne(nu).Exec(softdeletectx); err != nil {
			panic(err)
		}
		_, err = client.User.Query().Where(user.Name("u1")).First(softdeletectx)
		if err == nil {
			panic("found no delete user")
		} else {
			if !ent.IsNotFound(err) {
				panic(err)
			}
		}
	}
}
