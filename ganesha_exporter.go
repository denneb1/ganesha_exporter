package main

import (
	"github.com/denneb1/ganesha_exporter/dbus"
	"github.com/davecgh/go-spew/spew"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
)

func main() {
	var (
		listenAddress     = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":9587").String()
		metricsPath       = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		gandi             = kingpin.Flag("gandi", "Activate Gandi specific fields").Default("false").Bool()
		exporterCollector = kingpin.Flag("collector.exports", "Activate exports collector").Default("true").Bool()
	)
	ec := NewExportsCollector()
	var clientCollector = kingpin.Flag("collector.clients", "Activate clients collector").Default("true").Bool()
	cc := NewClientsCollector()

	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("ctld_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	dbus.Gandi = *gandi

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		prometheus.NewGoCollector(),
	)
	if *exporterCollector {
		reg.MustRegister(ec)
	}
	if *clientCollector {
		reg.MustRegister(cc)
	}
	http.Handle(*metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
			<head><title>ctld Exporter</title></head>
			<body>
			<h1>ctld Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
		if err != nil {
			log.Errorln(err)
		}
	})

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
	mgr := dbus.NewExportMgr()
	time, exports := mgr.ShowExports()
	spew.Dump(time, exports)
}
