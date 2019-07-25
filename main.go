package main

import (
	"fmt"
	"os"

	// "encoding/json"
	"net/http"
	// "errors"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

type Product struct {
	Name  string `json:"Name"`
	ID    string `json:"Id"`
	Count int    `json:"Count"`
}

type Inventory []Product

var MyInventory Inventory = Inventory{
	Product{Name: "Wrench", ID: "wrench", Count: 1},
	Product{Name: "Nalis", ID: "nails", Count: 1},
	Product{Name: "Hammer", ID: "hammer", Count: 1},
}

func generateRuntimeError(rw http.ResponseWriter, r *http.Request) {
	//Generate panic [runtime error: index out of range]
	fmt.Println(MyInventory[3].Name)
}

func handledSentryError(rw http.ResponseWriter, r *http.Request) {
	_, err := os.Open("filename.ext")
	if err != nil {
		sentry.CaptureException(err)
		//sentry.Flush(time.Second * 5)
	}

	// event := &sentry.NewEvent()
	// event.Message = "Hand-crafted event"
	// event.Extra["runtime.Version"] = runtime.Version()
	// event.Extra["runtime.NumCPU"] = runtime.NumCPU()

	// sentry.CaptureEvent(event)
}

func handleCheckout(rw http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(rw, "ParseForm() err: %v", err)
			return
		}
		fmt.Fprintf(rw, "whatever...")

	default:
		fmt.Fprintf(rw, "Sorry, only POST method is supported.")
	}
	// body_unicode = request.body.decode('utf-8')
	// order = json.loads(body_unicode)
	// cart = order['cart']
	// process_order(cart)
	// return Response(InventoryData)

}

// def process_order(cart):
//     global InventoryData
//     tempInventory = InventoryData
//     for item in cart:
//         itemID = item['id']
//         inventoryItem = find_in_inventory(itemID)
//         if inventoryItem['count'] <= 0:
//             raise Exception("Not enough inventory for " + itemID)
//         else:
//             inventoryItem['count'] -= 1
//             print( 'Success: ' + itemID + ' was purchased, remaining stock is ' + str(inventoryItem['count']) )
//     InventoryData = tempInventory

func main() {
	_ = sentry.Init(sentry.ClientOptions{
		Dsn: "https://a4efaa11ca764dd8a91d790c0926f810@sentry.io/1511084",
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			fmt.Println("************ Before Send ***********")
			if hint.Context != nil {
				if req, ok := hint.Context.Value(sentry.RequestContextKey).(*http.Request); ok {
					// You have access to the original Request
					transactionID := req.Header.Get("X-Transaction-ID")
					sessionID := req.Header.Get("X-Session-ID")

					sentry.ConfigureScope(func(scope *sentry.Scope) {
						//scope.SetUser(sentry.User{Email: "john.doe@example.com"})
						if len(transactionID) > 0 {
							fmt.Println("************ adding transaction ID:" + transactionID)
							scope.SetTag("transaction_id", transactionID)
						}
						if len(sessionID) > 0 {
							fmt.Println("&&&&&&& adding Session ID: " + sessionID)
							scope.SetTag("session-id", sessionID)
						}
					})

					//fmt.Println(req)
				}
			}
			//fmt.Println(event)
			return event
		},
		Debug:            true,
		AttachStacktrace: true,
	})

	sentryHandler := sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	})

	//http.Handle("/", sentryHandler.Handle(&handler{}))
	http.HandleFunc("/handled", sentryHandler.HandleFunc(handledSentryError))
	http.HandleFunc("/unhandled", sentryHandler.HandleFunc(generateRuntimeError))
	http.HandleFunc("/checkout", sentryHandler.HandleFunc(handleCheckout))

	fmt.Println("@@@@@@@@@Listening and serving HTTP on :3000")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		panic(err)
	}
}
