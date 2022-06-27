// Copyright (c) 2020 Khramtsov Aleksei (seniorGolang@gmail.com).
// This file (transport.go at 24.06.2020, 15:26) is subject to the terms and
// conditions defined in file 'LICENSE', which is part of this project source code.
package generator

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vetcher/go-astra"
	"github.com/vetcher/go-astra/types"

	"github.com/seniorGolang/tg/v2/pkg/tags"
)

const keyCode = "code"

const doNotEdit = "GENERATED BY 'T'ransport 'G'enerator. DO NOT EDIT."

const (
	tagLogger        = "log"
	tagDesc          = "desc"
	tagType          = "type"
	tagTag           = "tags"
	tagTests         = "tests"
	tagTrace         = "trace"
	tagFormat        = "format"
	tagSummary       = "summary"
	tagHandler       = "handler"
	tagExample       = "example"
	tagMetrics       = "metrics"
	tagUploadVars    = "http-upload"
	tagDownloadVars  = "http-download"
	tagHttpArg       = "http-args"
	tagHttpPath      = "http-path"
	tagDeprecated    = "deprecated"
	tagHttpPrefix    = "http-prefix"
	tagMethodHTTP    = "http-method"
	tagServerHTTP    = "http-server"
	tagHttpHeader    = "http-headers"
	tagHttpCookies   = "http-cookies"
	tagHttpSuccess   = "http-success"
	tagServerJsonRPC = "jsonRPC-server"
	tagHttpResponse  = "http-response"
	tagPackageJSON   = "packageJSON"
	tagPackageUUID   = "uuidPackage"
	tagSwaggerTags   = "swaggerTags"
	tagSwaggerDeep   = "swaggerDeep"
)

type Transport struct {
	hasJsonRPC bool
	version    string
	tags       tags.DocTags
	log        logrus.FieldLogger
	services   map[string]*service
}

func NewTransport(log logrus.FieldLogger, version, svcDir string, options ...Option) (tr Transport, err error) {

	tr.log = log
	tr.version = version
	tr.services = make(map[string]*service)

	var files []os.FileInfo
	if files, err = ioutil.ReadDir(svcDir); err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") {
			continue
		}
		var serviceAst *types.File
		svcDir, _ = filepath.Abs(svcDir)
		filePath := path.Join(svcDir, file.Name())
		if serviceAst, err = astra.ParseFile(filePath); err != nil {
			return
		}
		tr.tags = tr.tags.Merge(tags.ParseTags(serviceAst.Docs))
		for _, iface := range serviceAst.Interfaces {
			if len(tags.ParseTags(iface.Docs)) != 0 {
				service := newService(log, &tr, filePath, iface, options...)
				tr.services[iface.Name] = service

				if service.tags.Contains(tagServerJsonRPC) {
					tr.hasJsonRPC = true
				}
			}
		}
	}
	return
}

func (tr Transport) RenderAzure(appName, routePrefix, outDir, logLevel string, enableHealth bool) (err error) {
	return newAzure(&tr).render(appName, routePrefix, outDir, logLevel, enableHealth)
}

func (tr Transport) RenderSwagger(outDir string) (err error) {
	return newSwagger(&tr).render(outDir)
}

func (tr Transport) serviceKeys() (keys []string) {

	for serviceName := range tr.services {
		keys = append(keys, serviceName)
	}
	sort.Strings(keys)
	return
}

func (tr Transport) RenderClient(outDir string) (err error) {

	tr.cleanup(outDir)
	if err = os.MkdirAll(outDir, 0777); err != nil {
		return
	}

	if tr.hasTrace() {
		showError(tr.log, tr.renderClientTracer(outDir), "renderHTTP")
	}
	showError(tr.log, tr.renderClientOptions(outDir), "renderHTTP")
	if tr.hasJsonRPC {
		showError(tr.log, tr.renderClientJsonRPC(outDir), "renderHTTP")
	}
	for _, serviceName := range tr.serviceKeys() {
		svc := tr.services[serviceName]
		showError(tr.log, svc.renderClient(outDir), "renderHTTP")
	}
	return
}

func (tr Transport) RenderServer(outDir string) (err error) {

	tr.cleanup(outDir)

	if err = os.MkdirAll(outDir, 0777); err != nil {
		return
	}

	hasTrace := tr.hasTrace()
	hasMetric := tr.hasMetrics()

	showError(tr.log, tr.renderHTTP(outDir), "renderHTTP")
	showError(tr.log, tr.renderFiber(outDir), "renderFiber")
	showError(tr.log, tr.renderHeader(outDir), "renderHeader")
	showError(tr.log, tr.renderErrors(outDir), "renderErrors")
	showError(tr.log, tr.renderServer(outDir), "renderServer")
	showError(tr.log, tr.renderVersion(outDir), "renderVersion")
	showError(tr.log, tr.renderOptions(outDir), "renderOptions")
	if hasMetric {
		showError(tr.log, tr.renderMetrics(outDir), "renderMetrics")
	}
	if hasTrace {
		showError(tr.log, tr.renderTracer(outDir), "renderTracer")
	}
	if tr.hasJsonRPC {
		showError(tr.log, tr.renderJsonRPC(outDir), "renderJsonRPC")
	}

	for _, serviceName := range tr.serviceKeys() {
		svc := tr.services[serviceName]
		err = svc.render(outDir)
	}
	return
}

func (tr Transport) hasTrace() (hasTrace bool) {
	for _, serviceName := range tr.serviceKeys() {
		svc := tr.services[serviceName]
		if svc.tags.IsSet(tagTrace) {
			return true
		}
	}
	return
}

func (tr Transport) hasMetrics() (hasMetric bool) {
	for _, serviceName := range tr.serviceKeys() {
		svc := tr.services[serviceName]
		if svc.tags.IsSet(tagMetrics) {
			return true
		}
	}
	return
}

func showError(log logrus.FieldLogger, err error, msg string) {
	if err != nil {
		log.WithError(err).Error(msg)
	}
}
