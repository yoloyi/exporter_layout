package main

import (
	"context"
	"exporter_layout/collector"
	"github.com/prometheus/client_golang/prometheus"
	promcollectors "github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var (
	logLevel, cfgFile string
	cmd               = &cobra.Command{
		Use:                "scan",
		Short:              "scan",
		Long:               "",
		Example:            "scan",
		Version:            "0.0.1",
		PreRun:             cmdPreRunFunc,
		Run:                cmdRunFunc,
		FParseErrWhitelist: cobra.FParseErrWhitelist{},
		CompletionOptions:  cobra.CompletionOptions{},
	}
	listenAddress, metricsPath string
	disableExporterMetrics     bool
	maxProcs                   int
	errChan                    = make(chan error)
)

func initialize() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetReportCaller(true)
	l, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(l)
	log.Info("log.level：", logLevel)
	log.Info("config.path：", cfgFile)
}

func main() {
	cmd.Flags().StringVar(&logLevel, "log.level", "info", "日志等级.")
	cmd.Flags().StringVar(&cfgFile, "config.path", "/etc/scan/config.yml", "配置文件路径")

	cmd.Flags().StringVar(&listenAddress, "web.listen-address", ":9108", "Address to listen on for web interface and telemetry.")
	cmd.Flags().StringVar(&metricsPath, "web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	cmd.Flags().IntVar(&maxProcs, "runtime.gomaxprocs", runtime.NumCPU(), "The target number of CPUs Go will run on (GOMAXPROCS)")
	cmd.Flags().BoolVar(&disableExporterMetrics, "web.disable-exporter-metrics", true, "Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).")

	cobra.OnInitialize(initialize)

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

func cmdPreRunFunc(cmd *cobra.Command, args []string) {

}

func cmdRunFunc(cmd *cobra.Command, args []string) {
	runtime.GOMAXPROCS(maxProcs)
	log.Debug("msg", "Go MAXPROCS", "procs", runtime.GOMAXPROCS(0))
	serverHttp()
}

func newHandler() http.Handler {
	reg := prometheus.NewRegistry()
	reg.MustRegister(collector.NewSCollector())
	var handlerfor http.Handler
	if !disableExporterMetrics {
		reg.MustRegister(
			promcollectors.NewProcessCollector(promcollectors.ProcessCollectorOpts{}),
			promcollectors.NewGoCollector(),
		)
		handlerfor = promhttp.InstrumentMetricHandler(reg, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	} else {
		handlerfor = promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	}
	return handlerfor

}

func serverHttp() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Http Exporter</title></head>
             <body>
			 <h3>Deepfos Monitor</h3>
             <p><a href='` + metricsPath + `'>Metrics</a></p>
             <p><a href='/-/ready'>Health</a></p>
             </body>
             </html>`))
	})
	handlerfor := newHandler()
	mux.Handle(metricsPath, handlerfor)

	mux.HandleFunc("/-/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready"))
	})

	mux.HandleFunc("/-/reload", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("reload"))
	})

	server := &http.Server{Addr: listenAddress, Handler: mux}

	log.Info("Listening on address：", listenAddress)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()
	go listenSignal(server)
	log.Fatalln(<-errChan)
}

func listenSignal(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-sigs:
		log.Debug("Notify Shutdown")
		server.Shutdown(ctx)
		log.Debug("Shutdown")
	}
}
