package mevrelic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/justjack1521/mevconn"
	"github.com/newrelic/go-agent/v3/integrations/logcontext-v2/nrlogrus"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"time"
)

type NewRelic struct {
	LicenseKey  string
	EntityGUID  string
	EntityName  string
	Application *newrelic.Application
	client      *http.Client
}

func (a *NewRelic) Levels() []logrus.Level {
	return logrus.AllLevels
}

func NewRelicApplication() (*NewRelic, error) {
	config, err := mevconn.CreateNewRelicConfig()
	if err != nil {
		return nil, err
	}
	relic, err := newrelic.NewApplication(
		newrelic.ConfigAppName(config.ApplicationName()),
		newrelic.ConfigLicense(config.LicenseKey()),
		newrelic.ConfigAppLogDecoratingEnabled(true),
		newrelic.ConfigAppLogForwardingEnabled(false),
		func(cfg *newrelic.Config) {
			cfg.ErrorCollector.RecordPanics = true
		},
	)
	if err != nil {
		return nil, err
	}

	var client = &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   3 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			IdleConnTimeout:       90 * time.Second,
			MaxIdleConns:          10,
			MaxConnsPerHost:       10,
		},
		Timeout: 8 * time.Second,
	}

	return &NewRelic{
		LicenseKey:  config.LicenseKey(),
		EntityGUID:  config.ApplicationGUID(),
		EntityName:  config.ApplicationName(),
		Application: relic,
		client:      client,
	}, nil
}

func (a *NewRelic) Attach(logger *logrus.Logger) {
	nrl := nrlogrus.NewFormatter(a.Application, &logrus.TextFormatter{})
	logger.SetFormatter(nrl)
	logger.AddHook(a)
}

func (a *NewRelic) Fire(entry *logrus.Entry) error {
	evt := map[string]interface{}{
		"timestamp":   entry.Time.Unix(),
		"message":     entry.Message,
		"logtype":     entry.Level.String(),
		"entity.name": a.EntityName,
		"entity.guid": a.EntityGUID,
	}
	for k, v := range entry.Data {
		evt[k] = v
	}
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://log-api.newrelic.com/log/v1?Api-Key=%s", a.LicenseKey),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}
