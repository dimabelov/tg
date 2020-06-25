// Copyright (c) 2020 Khramtsov Aleksei (contact@altsoftllc.com).
// This file (transport-metrics.go at 14.05.2020, 2:13) is subject to the terms and
// conditions defined in file 'LICENSE', which is part of this project source code.
package generator

import (
	"path"
	"path/filepath"

	. "github.com/dave/jennifer/jen"
)

func (tr Transport) renderMetrics(outDir string) (err error) {

	srcFile := newSrc(filepath.Base(outDir))

	srcFile.PackageComment("GENERATED BY i2s. DO NOT EDIT.")

	srcFile.ImportAlias(packageKitPrometheus, "kitPrometheus")
	srcFile.ImportAlias(packageStdPrometheus, "stdPrometheus")

	srcFile.ImportName(packageFastHttp, "fasthttp")
	srcFile.ImportName(packageGoKitMetrics, "metrics")
	srcFile.ImportName(packageGoKitEndpoint, "endpoint")
	srcFile.ImportName(packagePrometheusHttp, "promhttp")
	srcFile.ImportName(packageFastHttpAdapt, "fasthttpadaptor")

	srcFile.Line().Add(prometheusCounterRequestCount())
	srcFile.Line().Add(prometheusCounterRequestCountAll())
	srcFile.Line().Add(prometheusSummaryRequestCount())

	srcFile.Line().Add(tr.serveMetricsFunc())

	return srcFile.Save(path.Join(outDir, "metrics.go"))
}

func (tr Transport) serveMetricsFunc() Code {

	return Func().Params(Id("srv").Op("*").Id("Server")).Id("ServeMetrics").Params(Id("address").String()).Block(

		Line().Id("srv").Dot("srvMetrics").Op("=").Op("&").Qual(packageFastHttp, "Server").Values(Dict{
			Id("ReadTimeout"): Qual(packageTime, "Second").Op("*").Lit(10),
			Id("Handler"):     Qual(packageFastHttpAdapt, "NewFastHTTPHandler").Call(Qual(packagePrometheusHttp, "Handler").Call()),
		}),

		Line().Go().Func().Params().Block(
			Err().Op(":=").Id("srv").Dot("srvMetrics").Dot("ListenAndServe").Call(Id("address")),
			Id("ExitOnError").Call(Id("srv").Dot("log"), Err(), Lit("serve metrics on ").Op("+").Id("address")),
		).Call(),
	)
}

func prometheusCounterRequestCount() (code *Statement) {

	return Var().Id("RequestCount").Op("=").Qual(packageKitPrometheus, "NewCounterFrom").Call(Qual(packageStdPrometheus, "CounterOpts").Values(
		DictFunc(func(d Dict) {
			d[Id("Name")] = Lit("count")
			d[Id("Namespace")] = Lit("service")
			d[Id("Subsystem")] = Lit("requests")
			d[Id("Help")] = Lit("Number of requests received")
		}),
	), Index().String().Values(Lit("method"), Lit("service"), Lit("success")))
}

func prometheusCounterRequestCountAll() (code *Statement) {

	return Var().Id("RequestCountAll").Op("=").Qual(packageKitPrometheus, "NewCounterFrom").Call(Qual(packageStdPrometheus, "CounterOpts").Values(
		DictFunc(func(d Dict) {
			d[Id("Name")] = Lit("all_count")
			d[Id("Namespace")] = Lit("service")
			d[Id("Subsystem")] = Lit("requests")
			d[Id("Help")] = Lit("Number of all requests received")
		}),
	), Index().String().Values(Lit("method"), Lit("service")))
}

func prometheusSummaryRequestCount() (code *Statement) {

	return Var().Id("RequestLatency").Op("=").Qual(packageKitPrometheus, "NewSummaryFrom").Call(Qual(packageStdPrometheus, "SummaryOpts").Values(
		DictFunc(func(d Dict) {
			d[Id("Name")] = Lit("latency_microseconds")
			d[Id("Namespace")] = Lit("service")
			d[Id("Subsystem")] = Lit("requests")
			d[Id("Help")] = Lit("Total duration of requests in microseconds")
		}),
	), Index().String().Values(Lit("method"), Lit("service"), Lit("success")))
}