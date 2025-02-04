package main

import (
	"html/template"
	"net/http"

	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"popkat/api"
	"popkat/constants"
	"popkat/state"
	"popkat/types"

	"popkat/routes/objects"

	"github.com/cloudflare/tableflip"
	docs "github.com/infinitybotlist/eureka/doclib"
	"github.com/infinitybotlist/eureka/jsonimpl"
	"github.com/infinitybotlist/eureka/uapi"
	"github.com/infinitybotlist/eureka/zapchi"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "embed"
)

//go:embed data/docs.html
var docsHTML string

var openapi []byte

// Simple middleware to handle CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// limit body to 10mb
		r.Body = http.MaxBytesReader(w, r.Body, 50*1024*1024)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "X-Client, Content-Type, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "X-Session-Invalid, Retry-After")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")

		if r.Method == "OPTIONS" {
			w.Write([]byte{})
			return
		}

		w.Header().Set("Content-Type", "application/json")

		next.ServeHTTP(w, r)
	})
}

func main() {
	state.Setup()

	var err error

	docs.DocsSetupData = &docs.SetupData{
		URL:         "https://popkat.purrquinox.com/",
		ErrorStruct: types.ApiError{},
		Info: docs.Info{
			Title:          "Popkat by Purrquinox",
			TermsOfService: "https:purrquinox.com/legal/terms",
			Version:        "2.0",
			Description:    "",
			Contact: docs.Contact{
				Name:  "Purrquinox",
				URL:   "https://purrquinox.com",
				Email: "support@purrquinox.com",
			},
			License: docs.License{
				Name: "MIT",
				URL:  "https://opensource.org/licenses/MIT",
			},
		},
	}

	docs.Setup()
	api.Setup()

	r := chi.NewRouter()

	r.Use(
		middleware.Recoverer,
		middleware.RealIP,
		middleware.CleanPath,
		corsMiddleware,
		zapchi.Logger(state.Logger, "api"),
		middleware.Timeout(30*time.Second),
	)

	routers := []uapi.APIRouter{
		&objects.Router{},
	}

	for _, router := range routers {
		name, desc := router.Tag()
		if name != "" {
			docs.AddTag(name, desc)
			uapi.State.SetCurrentTag(name)
		} else {
			panic("Router tag name cannot be empty")
		}

		router.Routes(r)
	}

	r.Get("/openapi", func(w http.ResponseWriter, r *http.Request) {
		w.Write(openapi)
	})

	docsTempl := template.Must(template.New("docs").Parse(docsHTML))

	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		docsTempl.Execute(w, map[string]string{
			"url": "/openapi",
		})
	})

	// Load openapi here to avoid large marshalling in every request
	openapi, err = jsonimpl.Marshal(docs.GetSchema())

	if err != nil {
		panic(err)
	}

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(constants.EndpointNotFound))
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(constants.MethodNotAllowed))
	})

	// If GOOS is windows, do normal http server
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		upg, _ := tableflip.New(tableflip.Options{})
		defer upg.Stop()

		go func() {
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGHUP)
			for range sig {
				state.Logger.Info("Received SIGHUP, upgrading server")
				upg.Upgrade()
			}
		}()

		// Listen must be called before Ready
		ln, err := upg.Listen("tcp", state.Config.Server.Port)

		if err != nil {
			state.Logger.Fatal("Error binding to socket", zap.Error(err))
		}

		defer ln.Close()

		server := http.Server{
			ReadTimeout: 30 * time.Second,
			Handler:     r,
		}

		go func() {
			err := server.Serve(ln)
			if err != http.ErrServerClosed {
				state.Logger.Error("Server failed due to unexpected error", zap.Error(err))
			}
		}()

		if err := upg.Ready(); err != nil {
			state.Logger.Fatal("Error calling upg.Ready", zap.Error(err))
		}

		<-upg.Exit()
	} else {
		// Tableflip not supported
		state.Logger.Warn("Tableflip not supported on this platform, this is not a production-capable server.")
		err = http.ListenAndServe(state.Config.Server.Port, r)

		if err != nil {
			state.Logger.Fatal("Error binding to socket", zap.Error(err))
		}
	}
}
