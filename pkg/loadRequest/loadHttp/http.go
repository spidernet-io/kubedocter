// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package loadHttp

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/spidernet-io/spiderdoctor/pkg/k8s/apis/system/v1beta1"
	config "github.com/spidernet-io/spiderdoctor/pkg/types"
)

type HttpMethod string

const (
	HttpMethodGet     = HttpMethod("GET")
	HttpMethodPost    = HttpMethod("POST")
	HttpMethodPut     = HttpMethod("PUT")
	HttpMethodDelete  = HttpMethod("DELETE")
	HttpMethodConnect = HttpMethod("CONNECT")
	HttpMethodOptions = HttpMethod("OPTIONS")
	HttpMethodPatch   = HttpMethod("PATCH")
	HttpMethodHead    = HttpMethod("HEAD")
)

type HttpRequestData struct {
	Method              HttpMethod
	Url                 string
	Qps                 int
	PerRequestTimeoutMS int
	RequestTimeSecond   int
	Header              map[string]string
	Http2               bool
	DisableKeepAlives   bool
	DisableCompression  bool
}

func HttpRequest(logger *zap.Logger, reqData *HttpRequestData) *v1beta1.HttpMetrics {
	logger.Sugar().Infof("http request=%v", reqData)
	req, _ := http.NewRequest(string(reqData.Method), reqData.Url, nil)
	duration := time.Duration(reqData.RequestTimeSecond) * time.Second
	for k, v := range reqData.Header {
		req.Header.Set(k, v)
	}

	logger.Sugar().Infof("http request Concurrency=%d", config.AgentConfig.Configmap.NethttpDefaultConcurrency)

	w := &Work{
		Request:            req,
		Concurrency:        config.AgentConfig.Configmap.NethttpDefaultConcurrency,
		QPS:                reqData.Qps,
		Timeout:            reqData.PerRequestTimeoutMS,
		DisableCompression: reqData.DisableCompression,
		DisableKeepAlives:  reqData.DisableKeepAlives,
		Http2:              reqData.Http2,
	}
	logger.Sugar().Infof("do http requests work=%v", w)
	w.Init()

	// The monitoring task timed out
	go func() {
		time.Sleep(duration)
		w.Stop()
	}()

	logger.Sugar().Infof("begin to request %v for duration %v ", w.Request.URL, duration.String())
	w.Run()
	logger.Sugar().Infof("finish all request %v for %s ", w.report.totalCount, w.Request.URL)
	// Collect metric reports
	metrics := w.AggregateMetric()
	return metrics
}
