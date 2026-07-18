package snyk

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

//
//import (
//	"fmt"
//	"net/http"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//)
//
//func TestUsers_GetCurrent(t *testing.T) {
//	setup()
//	defer teardown()
//
//	mux.HandleFunc("/user/me", func(w http.ResponseWriter, r *http.Request) {
//		assert.Equal(t, http.MethodGet, r.Method)
//		_, _ = fmt.Fprint(w, `
//{
//  "email":"test-user@snyk.io",
//  "id":"long-uuid",
//  "username":"test-user"
//}
//`)
//	})
//	expectedUser := &User{
//		Email:    "test-user@snyk.io",
//		ID:       "long-uuid",
//		Username: "test-user",
//	}
//
//	actualUser, _, err := client.Users.GetCurrent(ctx)
//
//	assert.NoError(t, err)
//	assert.Equal(t, expectedUser, actualUser)
//}

func TestUsers_GetSelf(t *testing.T) {
	setup(t)
	defer teardown()

	mux.HandleFunc("/self", func(w http.ResponseWriter, r *http.Request) {
		assertRequestAPIVersion(t, r, usersAPIVersion)
		_, _ = fmt.Fprint(w, `{
  "data":{"id":"user-id","type":"user","attributes":{"name":"Test User"}}
}`)
	})

	user, _, err := client.Users.GetSelf(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "user-id", user.ID)
}

//
//func TestUsers_Get(t *testing.T) {
//	setup()
//	defer teardown()
//
//	mux.HandleFunc("/user/long-uuid", func(w http.ResponseWriter, r *http.Request) {
//		assert.Equal(t, http.MethodGet, r.Method)
//		_, _ = fmt.Fprint(w, `
//{
//  "email":"test-user@snyk.io",
//  "id":"long-uuid",
//  "name": "Test User",
//  "username":"test-user"
//}
//`)
//	})
//	expectedUser := &User{
//		Email:    "test-user@snyk.io",
//		ID:       "long-uuid",
//		Name:     "Test User",
//		Username: "test-user",
//	}
//
//	actualUser, _, err := client.Users.Get(ctx, "long-uuid")
//
//	assert.NoError(t, err)
//	assert.Equal(t, expectedUser, actualUser)
//}
