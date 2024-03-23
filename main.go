package main

import (
	"encoding/json"
	"errors"
	"github.com/azinudinachzab/belajar-microservices/model"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func main() {
	//if err := godotenv.Load(); err != nil {
	//	log.Fatalf("%v\n", err)
	//}
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		// "True-Client-IP", "X-Forwarded-For", "X-Real-IP", "X-Request-Id",
		// "Origin", "Accept", "Content-Type", "Authorization", "Token"
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           86400,
	}))
	r.Use(httprate.LimitByIP(80, 1*time.Minute))
	r.Use(middleware.CleanPath)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.NotFound(Custom404)
	r.MethodNotAllowed(Custom405)

	// router
	svc, ok := os.LookupEnv("SERVICE")
	if !ok {
		log.Fatalf("service name not defined")
	}

	switch svc {
	case "customer":
		initCustomerController(r)
	case "order":
		initOrderController(r)
	case "payment":
		initPaymentController(r)
	default:
		log.Fatalf("service not registered")
		return
	}
	server := &http.Server{
		Addr:         os.Getenv("PORT"),
		Handler:      r,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("listen and serve returned err: %v", err)
	}
}

func Custom404(w http.ResponseWriter, _ *http.Request) {
	err := errors.New(model.ECodeNotFound + "route does not exist")
	w.WriteHeader(http.StatusNotFound)
	toJSON(w, err)
}

func Custom405(w http.ResponseWriter, _ *http.Request) {
	err := errors.New(model.ECodeMethodFail + "method is not valid")
	w.WriteHeader(http.StatusMethodNotAllowed)
	toJSON(w, err)
}

func toJSON(w http.ResponseWriter, data interface{}) {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("fail to encode to JSON", err)
	}
}

func initCustomerController(r *chi.Mux) {
	r.Get("/check-customer", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL)
		nama := r.URL.Query().Get("name")
		for idx, val := range []string{"steven", "william", "alexander", "jonathan"} {
			if val == nama {
				w.WriteHeader(http.StatusOK)
				toJSON(w, struct {
					CustomerID   int    `json:"customer_id"`
					Name         string `json:"name"`
					IsRegistered bool   `json:"is_registered"`
				}{
					CustomerID:   idx + 1,
					Name:         val,
					IsRegistered: true,
				})
				return
			}
		}
		err := errors.New(model.ECodeNotFound + "customer not found")
		w.WriteHeader(http.StatusBadRequest)
		toJSON(w, struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		})
	})
}

func initOrderController(r *chi.Mux) {
	r.Get("/check-order", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL)
		nama := r.URL.Query().Get("order_id")
		for _, val := range []string{"123", "456", "789"} {
			if val == nama {
				w.WriteHeader(http.StatusOK)
				toJSON(w, struct {
					OrderID  string `json:"order_id"`
					Name     string `json:"name"`
					Quantity int    `json:"quantity"`
				}{
					OrderID:  val,
					Name:     "Pulpen",
					Quantity: 10,
				})
				return
			}
		}
		err := errors.New(model.ECodeNotFound + "order not found")
		w.WriteHeader(http.StatusBadRequest)
		toJSON(w, struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		})
	})
}

func initPaymentController(r *chi.Mux) {
	r.Get("/do-payment", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL)
		custName := r.URL.Query().Get("name")
		orderID := r.URL.Query().Get("order_id")

		client := &http.Client{}
		req, err := http.NewRequest(http.MethodGet, os.Getenv("CUSTOMER_URL")+"/check-customer?name="+custName, nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			toJSON(w, struct {
				Error string `json:"error"`
			}{
				Error: err.Error(),
			})
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			toJSON(w, struct {
				Error string `json:"error"`
			}{
				Error: err.Error(),
			})
			return
		}

		defer resp.Body.Close()
		customer := struct {
			CustomerID   int    `json:"customer_id"`
			Name         string `json:"name"`
			IsRegistered bool   `json:"is_registered"`
		}{}
		if err := json.NewDecoder(resp.Body).Decode(&customer); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			toJSON(w, struct {
				Error string `json:"error"`
			}{
				Error: err.Error(),
			})
			return
		}
		if !customer.IsRegistered {
			err := errors.New(model.ECodeNotFound + "customer not found")
			w.WriteHeader(http.StatusBadRequest)
			toJSON(w, struct {
				Error string `json:"error"`
			}{
				Error: err.Error(),
			})
			return
		}

		req1, err := http.NewRequest(http.MethodGet, os.Getenv("ORDER_URL")+"/check-order?order_id="+orderID, nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			toJSON(w, struct {
				Error string `json:"error"`
			}{
				Error: err.Error(),
			})
			return
		}
		resp1, err := client.Do(req1)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			toJSON(w, struct {
				Error string `json:"error"`
			}{
				Error: err.Error(),
			})
			return
		}

		defer resp1.Body.Close()
		order := struct {
			OrderID  string `json:"order_id"`
			Name     string `json:"name"`
			Quantity int    `json:"quantity"`
		}{}
		if err := json.NewDecoder(resp1.Body).Decode(&order); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			toJSON(w, struct {
				Error string `json:"error"`
			}{
				Error: err.Error(),
			})
			return
		}
		if order.OrderID == "" {
			err := errors.New(model.ECodeNotFound + "order not found")
			w.WriteHeader(http.StatusBadRequest)
			toJSON(w, struct {
				Error string `json:"error"`
			}{
				Error: err.Error(),
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		toJSON(w, struct {
			CustomerName  string `json:"customer_name"`
			CustomerID    int    `json:"customer_id"`
			OrderID       string `json:"order_id"`
			OrderName     string `json:"order_name"`
			Qty           int    `json:"qty"`
			PaymentStatus string `json:"payment_status"`
		}{
			CustomerName:  customer.Name,
			CustomerID:    customer.CustomerID,
			OrderID:       order.OrderID,
			OrderName:     order.Name,
			Qty:           order.Quantity,
			PaymentStatus: "Success",
		})
	})
}
