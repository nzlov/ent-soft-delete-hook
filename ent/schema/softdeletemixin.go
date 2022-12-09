package schema

import (
	"context"
	"fmt"
	"time"

	gen "entgo.io/bug/ent"
	"entgo.io/ent/entql"
	"entgo.io/ent/schema"

	"entgo.io/bug/ent/hook"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type DeletedTimeAnnotation struct {
	OK bool
}

func (d DeletedTimeAnnotation) Name() string {
	return "DeletedTime"
}

type SoftDeleteMixin struct {
	mixin.Schema
}

func (SoftDeleteMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("deleted_time").Optional(),
	}
}

func (SoftDeleteMixin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		DeletedTimeAnnotation{OK: true},
	}
}

type softDeleteKey struct{}

// SkipSoftDelete returns a new context that skips the soft-delete interceptor/mutators.
func SkipSoftDelete(parent context.Context) context.Context {
	return context.WithValue(parent, softDeleteKey{}, true)
}

func (SoftDeleteMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		ent.TraverseFunc(func(ctx context.Context, q ent.Query) error {
			fmt.Println("Interceptor:")
			if skip, _ := ctx.Value(softDeleteKey{}).(bool); skip {
				fmt.Println("Interceptor:Skip")
				return nil
			}
			if f, ok := gen.Filter(q).(interface {
				WhereDeletedTime(p entql.TimeP)
			}); ok {
				f.WhereDeletedTime(entql.TimeNil())
				fmt.Println("Interceptor:delete")
			} else {
				fmt.Println("Interceptor:no delete")
			}
			return nil
		}),
	}
}

func (SoftDeleteMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(
			func(next ent.Mutator) ent.Mutator {
				fmt.Println("OnHook:")
				return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
					if skip, _ := ctx.Value(softDeleteKey{}).(bool); skip {
						return next.Mutate(ctx, m)
					}
					return gen.MarkAsDeleted(ctx, m, time.Now())
				})
			},
			ent.OpDeleteOne|ent.OpDelete,
		),
	}
}
