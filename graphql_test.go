package graphql

import (
	"bytes"
	"encoding/json"
	"testing"
)

type helloWorldResolver struct{}

func (r *helloWorldResolver) Hello() string {
	return "Hello world!"
}

type starWarsResolver struct{}

func (r *starWarsResolver) Hero() *userResolver {
	return &userResolver{
		id:   "2001",
		name: "R2-D2",
		friends: []*userResolver{
			{name: "Luke Skywalker"},
			{name: "Han Solo"},
			{name: "Leia Organa"},
		},
	}
}

type userResolver struct {
	id      string
	name    string
	friends []*userResolver
}

func (r *userResolver) ID() string {
	return r.id
}

func (r *userResolver) Name() string {
	return r.name
}

func (r *userResolver) Friends() []*userResolver {
	return r.friends
}

var tests = []struct {
	name     string
	schema   string
	resolver interface{}
	query    string
	result   string
}{
	{
		name: "HelloWorld",
		schema: `
			type Query {
				hello: String
			}
		`,
		resolver: &helloWorldResolver{},
		query: `
			{
				hello
			}
		`,
		result: `
			{
				"hello": "Hello world!"
			}
		`,
	},
	{
		name: "StarWars",
		schema: `
			type Query {
				hero: User
			}

			type User {
				id: String
				name: String
				friends: [User]
			}
		`,
		resolver: &starWarsResolver{},
		query: `
			{
				hero {
					id
					name
					friends {
						name
					}
				}
			}
		`,
		result: `
			{
				"hero": {
					"id": "2001",
					"name": "R2-D2",
					"friends": [
						{
							"name": "Luke Skywalker"
						},
						{
							"name": "Han Solo"
						},
						{
							"name": "Leia Organa"
						}
					]
				}
			}
		`,
	},
}

func TestAll(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema, err := NewSchema(test.schema, test.name, test.resolver)
			if err != nil {
				t.Fatal(err)
			}

			got, err := schema.Exec(test.query)
			if err != nil {
				t.Fatal(err)
			}

			want := formatJSON([]byte(test.result))
			if !bytes.Equal(got, want) {
				t.Logf("want: %s", want)
				t.Logf("got:  %s", got)
				t.Fail()
			}
		})
	}
}

func formatJSON(data []byte) []byte {
	var v interface{}
	json.Unmarshal(data, &v)
	b, _ := json.Marshal(v)
	return b
}
