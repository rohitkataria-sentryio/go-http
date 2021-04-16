package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

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
	"nails":  &Product{Name: "Nails", ID: "nails", Count: 1},
	"hammer": &Product{Name: "Hammer", ID: "hammer", Count: 1},
}

type handler struct{}

func (h *handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
		hub.Scope().SetExtra("extra_data_1", "some extra value")
		ipAddressVal := r.Header.Get("X-FORWARDED-FOR")
		if len(ipAddressVal) > 0 {
			hub.Scope().SetUser(sentry.User{IPAddress: ipAddressVal})
		}
		transactionID := r.Header.Get("X-Transaction-ID")
		if len(transactionID) > 0 {
			hub.Scope().SetTag("transaction_id", transactionID)
		}

	}

	endpointRequest := r.URL.Path
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category: "Network",
		Message:  "Processing request for endpoint: " + endpointRequest,
		Level:    sentry.LevelInfo,
	})

	switch endpointRequest {
	case "/unhandled":
		generateRuntimeError(rw, r)

	case "/handled":
		generateSentryError(rw, r)

	case "/message":
		sendSentryCaptureMessage(rw, r)

	case "/checkout":
		if r.Method == "POST" {
			handleCheckout(rw, r)
		} else {
			fmt.Fprintf(rw, "Endpoint supports only POST method")
		}

	case "/success":
		success(rw, r)

	case "/favicon.ico":
		http.ServeFile(rw, r, "static/favicon.ico")

	default:
		fmt.Fprintf(rw, "Welcome to Go...")
	}
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
		if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
			hub.CaptureException(err)
		}
	}
}

func sendSentryCaptureMessage(rw http.ResponseWriter, r *http.Request) {
	if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
		hub.CaptureMessage("Send this message to Sentry")
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
	userEmail := data["email"].(string)
	if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
		hub.Scope().SetUser(sentry.User{Email: userEmail})
		hub.Scope().SetExtra("Cart", data["cart"].([]interface{}))
	}

	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category: "Workflow",
		Message:  "Checkout cart for user " + userEmail,
		Level:    sentry.LevelInfo,
	})

	processOrder(data, rw)
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

func success(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("success"))
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	if DSN = os.Getenv("DSN"); DSN == "" {
		DSN = "https://b3d031eea696421a899b78549349e888@o87286.ingest.sentry.io/5722093"
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
		ServerName:       "ServerName",
		IgnoreErrors:     []string{"MyIOError", "MyDBError"},
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			fmt.Println(event.EventID)

			// Fingerprinting - this updates the fingerprint for the error produced by /handled endpoint
			// if ex, _ := hint.OriginalException.(OpenFileError); ex.Error() == "open filename.ext: no such file or directory" {
			// 	event.Fingerprint = []string{"open-file-error"}
			// }

			// if hint.Context != nil {
			// if req, ok := hint.Context.Value(sentry.RequestContextKey).(*http.Request); ok {
			// 	// You have access to the original Request object
			// 	fmt.Println(req)
			// }
			// }
			return event
		},
		// Specify either TracesSampleRate or set a TracesSampler to enable tracing
		// TracesSampleRate: 1.0,
		TracesSampler: sentry.TracesSamplerFunc(func(ctx sentry.SamplingContext) sentry.Sampled {
			// As an example, this custom sampler does not send some transactions to Sentry based on their name.
			hub := sentry.GetHubFromContext(ctx.Span.Context())
			name := hub.Scope().Transaction()
			fmt.Println("transaction:", name)

			if name == "GET /favicon.ico" {
				return sentry.SampledFalse
			}

			// ignore pre-flight requests
			if strings.HasPrefix(name, "OPTIONS") {
				return sentry.SampledFalse
			}

			// if strings.HasPrefix(name, "HEAD") {
			// 	return sentry.SampledFalse
			// }

			// As an example, sample some transactions with a uniform rate.
			// if strings.HasPrefix(name, "POST") {
			// 	return sentry.UniformTracesSampler(0.2).Sample(ctx)
			// }

			// On production, use TracesSampleRate with a rate adequate for your traffic,
			//  or use the SamplingContext to
			// customize sampling per-transaction.
			return sentry.SampledTrue
		}),
	})

	sentryHandler := sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	})
	c := cors.AllowAll()
	http.Handle("/", c.Handler(sentryHandler.Handle(c.Handler(&handler{}))))
	fmt.Println("Go Server listening on port 3000...")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		panic(err)
	}

}

type OpenFileError struct {
	Message string
}

func (e OpenFileError) Error() string {
	return "open filename.ext: no such file or directory"
	// return "OpenFileError: " + e.Message
}
