// Package polling handles the association between an address and a user's fitness data
package polling

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	geth "github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis"
	homedir "github.com/mitchellh/go-homedir"
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

// Testing variables
var (
	after = time.After
	done  = func() {}
)

func (s service) Poll(_ context.Context, a []byte) (uint64, uint64) {
	// TODO: Write Poll function
	return 0, 0
}

func (s service) Register(ctx context.Context, a []byte) string {
	// Create a Google OAuth2 configuration
	c, err := readOAuthFile()
	if err != nil {
		log.Fatalln("Cannot read OAuth config: " + err.Error())
	}
	c.Scopes = []string{fitness.FitnessActivityReadScope} // Request fitness activity from Google Fit

	// TODO: rewrite with a real server, in order to work with multiple pending requests
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

	c.RedirectURL = ts.URL

	// Routine that waits for the user's approval
	go func() {
		defer ts.Close()

		var code string
		select {
		case code = <-ch:
			log.Println("Code received")
		case <-after(60 * time.Second):
			log.Println("No answer received for 60 seconds")
			// Done testing method
			done()
			return
		}

		// Exchange the code for a token
		t, err := c.Exchange(ctx, code)
		if err != nil {
			log.Fatalf("Token exchange error: %v", err)
		}

		// Save token to redis database
		err = saveToken(s, a, t)
		if err != nil {
			log.Fatalf("Token save error: %v", err)
		}

		// Done testing method
		done()
	}()

	return c.AuthCodeURL(rs)
}

// readOAuthFile gets the ClientID and ClientSecret from a JSON file
func readOAuthFile() (*oauth2.Config, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadFile(home + "/.notafridge/fridgelethics/client_id.json")
	if err != nil {
		return nil, err
	}

	config, err := google.ConfigFromJSON(b)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// saveToken saves a OAuth2 token to a redis database
func saveToken(s service, a []byte, t *oauth2.Token) error {
	var ud UserData

	// Get current user data
	sa := geth.Bytes2Hex(a)
	cmd := redis.NewCmd("get", sa)
	s.db.Process(cmd)
	val, err := cmd.Result()
	if err != nil {
		ud = UserData{
			T: *t,
			D: 0,
			C: 0,
		}
	} else {
		err = ud.FromBytes([]byte(val.(string)))
		if err != nil {
			log.Fatalln(err)
		}
		ud.T = *t
	}

	b, err := ud.ToBytes()
	if err != nil {
		log.Fatalln(err)
	}

	err = s.db.Set(sa, b, 0).Err()
	if err != nil {
		return err
	}

	return nil
}
