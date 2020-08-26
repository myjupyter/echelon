package store

import (
	"context"

	log "github.com/sirupsen/logrus"
	ttool "github.com/tarantool/go-tarantool"

	"github.com/volatiletech/authboss/v3"
)

const (
	TableName = "users"
)

type TarantoolStorer struct {
	connection *ttool.Connection
}

func NewTarantoolStorer(addr string, opts ttool.Opts) *TarantoolStorer {
	conn, err := ttool.Connect(addr, opts)
	if err != nil {
		log.Fatal(err)
	}
	return &TarantoolStorer{connection: conn}
}

func (ts *TarantoolStorer) Save(_ context.Context, user authboss.User) error {
	u := user.(*User)
	log.Info(u)
	_, err := ts.connection.Replace(TableName, []interface{}{u.ID, u.Name, u.Email, u.Password, u.RoleID})
	if err != nil {
		return err
	}
	return nil
}

func (ts *TarantoolStorer) Load(_ context.Context, key string) (authboss.User, error) {
	resp, err := ts.connection.Call("box.space."+TableName+".index.by_email:select", []interface{}{key})
	if err != nil || len(resp.Data[0].([]interface{})) == 0 {
		log.Error(resp.String())
		return nil, authboss.ErrUserNotFound
	}
	var user authboss.User
	user = &User{
		ID:        int(resp.Data[0].([]interface{})[0].(uint64)),
		Name:      resp.Data[0].([]interface{})[1].(string),
		Email:     resp.Data[0].([]interface{})[2].(string),
		Password:  resp.Data[0].([]interface{})[3].(string),
		RoleID:    resp.Data[0].([]interface{})[4].(string),
		Confirmed: true,
	}
	return user, nil
}

func (ts *TarantoolStorer) New(_ context.Context) authboss.User {
	return &User{RoleID: "user"}
}

func (ts *TarantoolStorer) Create(ctx context.Context, user authboss.User) error {
	return ts.Save(ctx, user)
}
