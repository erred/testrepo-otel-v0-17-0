package main

import (
	"context"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.seankhliao.com/testrepo-otel-v0-17-0/internal/setup"
)

func main() {
	ctx := context.Background()
	// shutdown, err := stdout.InstallNewPipeline(nil, nil)
	_, err := setup.InstallOtlpPipeline(ctx)
	if err != nil {
		log.Fatal(err)
	}

	tracer := otel.Tracer(os.Args[1])

	switch os.Args[1] {
	case "a":
		// pprof
		go http.ListenAndServe(":8080", nil)

		for range time.NewTicker(time.Second).C {
			func() {
				ctx := context.Background()
				ctx, span := tracer.Start(ctx, "ping")
				defer span.End()

				u := "http://svcb.default.svc"
				res, err := otelhttp.Get(ctx, u)
				if err != nil {
					log.Printf("traceid=%s err=%q", span.SpanContext().TraceID, err.Error())
					return
				} else if res.StatusCode != 200 {
					log.Printf("traceid=%s status=%q", span.SpanContext().TraceID, res.Status)
					return
				}
				defer res.Body.Close()
				b, err := io.ReadAll(res.Body)
				if err != nil {
					log.Printf("traceid=%s err=%q", span.SpanContext().TraceID, err.Error())
					return
				}
				log.Printf("traceid=%s msg=%s", span.SpanContext().TraceID, string(b))
			}()
		}
	case "b":
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			res, err := otelhttp.Get(ctx, "http://svcc.default.svc")
			if err != nil {
				log.Println(err)
				return
			}
			defer res.Body.Close()
			io.Copy(w, res.Body)
		})

		http.ListenAndServe(":8080", otelhttp.NewHandler(http.DefaultServeMux, "handler"))
	case "c":
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			_, span := tracer.Start(r.Context(), "pong")
			defer span.End()

			w.Write([]byte("pog"))
		})

		http.ListenAndServe(":8080", otelhttp.NewHandler(http.DefaultServeMux, "handler"))
	}
}
