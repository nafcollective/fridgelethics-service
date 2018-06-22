// Package polling handles the association between an address and a user's fitness data
package polling // import "github.com/nafcollective/fridgelethics-service/polling/v0"

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	fitness "google.golang.org/api/fitness/v1"
)

type Service interface {
	Poll(context.Context, []byte) (uint64, uint64)
	Register(context.Context, []byte) string
}

type service struct {
	db *redis.Client
	fs *fitness.Service
}

func (s service) Poll(_ context.Context, a []byte) (uint64, uint64) {
	return 0, 0
}

func (s service) Register(ctx context.Context, a []byte) string {
	urlCh := make(chan string)
	go doOauth(ctx, s, a, urlCh)
	return <-urlCh
}

func NewService(o *redis.Options) (Service, error) {
	s := &service{
		db: redis.NewClient(o),
		fs: nil,
	}

	_, err := s.db.Ping().Result()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func doOauth(ctx context.Context, s service, a []byte, urlCh chan string) {
	c := &oauth2.Config{
		ClientID:     "", // TODO
		ClientSecret: "", // TODO
		Endpoint:     google.Endpoint,
		Scopes:       []string{fitness.FitnessActivityReadScope},
	}

	ch := make(chan string)
	rs := fmt.Sprintf("st%d", time.Now().UnixNano())
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/favicon.ico" {
			http.Error(rw, "", 404)
			return
		}
		if req.FormValue("state") != rs {
			log.Printf("State doesn't match: req = %#v", req)
			http.Error(rw, "", 500)
			return
		}
		if code := req.FormValue("code"); code != "" {
			fmt.Fprintf(rw, "Redirecting...")
			rw.(http.Flusher).Flush()
			ch <- code
			return
		}
		log.Printf("no code")
		http.Error(rw, "", 500)
	}))
	defer ts.Close()

	c.RedirectURL = ts.URL
	urlCh <- c.AuthCodeURL(rs)
	code := <-ch

	t, err := c.Exchange(ctx, code)
	if err != nil {
		log.Fatalf("Token exchange error: %v", err)
	}

	err = saveToken(s, a, t)
	if err != nil {
		log.Fatalf("Token save error: %v", err)
	}
}

func saveToken(s service, a []byte, t *oauth2.Token) error {
	var ud UserData

	sa := common.Bytes2Hex(a)
	cmd := redis.NewCmd("get", sa)
	s.db.Process(cmd)
	val, err := cmd.Result()
	if err != nil {
		ud = UserData{
			T: t,
			D: 0,
			C: 0,
		}
	} else {
		ud = val.(UserData)
		ud.T = t
	}

	err = s.db.Set(sa, ud, 0).Err()
	if err != nil {
		return err
	}

	// TODO

	return nil
}
