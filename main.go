package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/ulule/deepcopier"
)

var (
	DSN string
)

// test suspect commits 1

//Product in the Inventory
type Product struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	Count int    `json:"Count"`
}

// MyInventory correlates with the REACT store demo inventory
var MyInventory = map[string]*Product{
	"wrench": &Product{Name: "Wrench", ID: "wrench", Count: 1},
	"nails":  &Product{Name: "Nalis", ID: "nails", Count: 1},
	"hammer": &Product{Name: "Hammer", ID: "hammer", Count: 1},
}

func generateRuntimeError(rw http.ResponseWriter, r *http.Request) {
	//Generate panic [runtime error: index out of range]
	productIDs := []string{}

	for key := range MyInventory {
		productIDs = append(productIDs, key)
	}
	fmt.Println(productIDs[4])
}

func generateSentryError(rw http.ResponseWriter, r *http.Request) {
	_, err := os.Open("filename.ext")
	if err != nil {
		sentry.CaptureException(err)
	}
}

func handleCheckout(rw http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(rw, err.Error(), 500)
	}

	var data map[string]interface{}
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		http.Error(rw, err.Error(), 500)
	}

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{Email: data["email"].(string)})
		scope.SetExtra("Cart", data["cart"].([]interface{}))
		processOrder(data, rw)
	})
}

func processOrder(data map[string]interface{}, rw http.ResponseWriter) error {
	tmpInventory := make(map[string]*Product)
	for k, p := range MyInventory {
		tmpPrd := &Product{}
		deepcopier.Copy(p).To(tmpPrd)
		tmpInventory[k] = tmpPrd
	}

	cart := data["cart"].([]interface{})
	var currentPrdID string
	for _, value := range cart {
		nestedMap, ok := value.(map[string]interface{})
		if ok {
			currentPrdID = nestedMap["id"].(string)
			if tmpInventory[currentPrdID].Count == 0 {
				panic("Not enough inventory for " + currentPrdID)
			} else {
				tmpInventory[currentPrdID].Count--
			}
		}
	}
	return nil
}

func routeRequest(rw http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {

	case "/unhandled":
		generateRuntimeError(rw, r)

	case "/handled":
		generateSentryError(rw, r)

	case "/checkout":
		if r.Method == "POST" {
			handleCheckout(rw, r)
		} else {
			fmt.Fprintf(rw, "Endpoint supports ony POST method")
		}

	default:
		fmt.Fprintf(rw, "Welcome to Go...")
	}
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	if DSN = os.Getenv("DSN"); DSN == "" {
		DSN = "https://a4efaa11ca764dd8a91d790c0926f810@sentry.io/1511084"
	}
	fmt.Println("DSN", DSN)
}

func main() {
	_ = sentry.Init(sentry.ClientOptions{
		Dsn:              DSN,
		Release:          os.Args[1],
		Environment:      "prod",
		Debug:            true,
		AttachStacktrace: true,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			if hint.Context != nil {
				if req, ok := hint.Context.Value(sentry.RequestContextKey).(*http.Request); ok {
					// You have access to the original Request here
					sentry.ConfigureScope(func(scope *sentry.Scope) {
						transactionID := req.Header.Get("X-Transaction-ID")
						sessionID := req.Header.Get("X-Session-ID")
						if len(transactionID) > 0 {
							scope.SetTag("transaction_id", transactionID)
						}
						if len(sessionID) > 0 {
							scope.SetTag("session-id", sessionID)
						}
					})
				}
			}
			log.Printf("BeforeSend event [%s]", event.EventID)
			return event
		},
		// Specify either TracesSampleRate or set a TracesSampler to enable tracing.
		TracesSampleRate: 1.0,
		/*TracesSampler: sentry.TracesSamplerFunc(func(ctx sentry.SamplingContext) sentry.Sampled {
			// As an example, this custom sampler does not send some
			// transactions to Sentry based on their name.
			hub := sentry.GetHubFromContext(ctx.Span.Context())
			name := hub.Scope().Transaction()
			if name == "GET /favicon.ico" {
				return sentry.SampledFalse
			}
			// if strings.HasPrefix(name, "HEAD") {
			// 	return sentry.SampledFalse
			// }

			// As an example, sample some transactions with a
			// uniform rate.
			// if strings.HasPrefix(name, "POST") {
			// 	return sentry.UniformTracesSampler(0.2).Sample(ctx)
			// }

			// Sample all other transactions for testing. On
			// production, use TracesSampleRate with a rate adequate
			// for your traffic, or use the SamplingContext to
			// customize sampling per-transaction.
			return sentry.SampledTrue
		}),*/
	})

	sentryHandler := sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	})

	c := cors.AllowAll()
	handler := sentryHandler.HandleFunc(routeRequest)

	fmt.Println("Go Server listening on port 3000...")

	if err := http.ListenAndServe(":3000", c.Handler(handler)); err != nil {
		panic(err)
	}

}
