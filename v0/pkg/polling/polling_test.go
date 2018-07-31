package polling_test

import (
	"context"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	. "github.com/nafcollective/fridgelethics-service/v0/pkg/polling"
	"github.com/skratchdot/open-golang/open"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	geth "github.com/ethereum/go-ethereum/common"
)

var _ = Describe("Polling", func() {
	var (
		svc     Service
		timeout chan time.Time
		done    chan bool
	)

	BeforeSuite(func() {
		timeout = make(chan time.Time)
		SetAfter(func(time.Duration) <-chan time.Time {
			return timeout
		})

		done = make(chan bool)
		SetDone(func() {
			done <- true
		})
	})

	BeforeEach(func() {
		ctx := context.Background()
		errc := make(chan error)
		var err error
		svc, err = NewService(ctx, &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}, errc)
		Ω(err).Should(BeNil())
	})

	Describe("Registering a user", func() {
		Context("When registering a new user", func() {
			It("Should return a URL", func() {
				url := svc.Register(context.Background(), geth.HexToAddress("0x1").Bytes())
				Ω(url).ShouldNot(BeNil())
				Ω(url).ShouldNot(BeEmpty())
			})

			PIt("Should Timeout after 60 seconds", func() {
				svc.Register(context.Background(), geth.HexToAddress("0x1").Bytes())

				timeout <- time.Now()
				<-done

				// TODO: Rewrite the next bits
				res, err := http.Get("http://localhost/")
				Ω(err).ShouldNot(BeNil())
				Ω(res.Body).Should(BeNil())
			})

			PIt("Should receive and save a token", func() {
				url := svc.Register(context.Background(), geth.HexToAddress("0x1").Bytes())

				err := open.Run(url)
				Ω(err).Should(BeNil())

				<-done

				// TODO: Complete this test
			})
		})
	})
})
