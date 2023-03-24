package config

import (
	"crypto/tls"
	"fmt"
	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"net"
	"net/http"
	"time"
)

// Scraper TODO change yaml.Node fields to correct types
type Scraper struct {
	Name                  string `yaml:"Name"`
	AppsJSONPath          string `yaml:"AppsJSONPath"`
	TimeoutSeconds        string `yaml:"TimeoutSeconds"`
	LoadingTimeoutSeconds string `yaml:"LoadingTimeoutSeconds"`
	JSON                  string `yaml:"JSON"`
	MaxDepth              string `yaml:"MaxDepth"`
	//visitedLinks           string `yaml:"VisitedLinks"`
	MaxVisitedLinks        string           `yaml:"MaxVisitedLinks"`
	MsDelayBetweenRequests string           `yaml:"MsDelayBetweenRequests"`
	UserAgent              string           `yaml:"UserAgent"`
	Collector              *colly.Collector `yaml:"-"`
	//componentConfigs       ComponentConfigs
}

type CrawlConfigMap map[string]*Scraper

func (c *Scraper) collector() (*colly.Collector, error) {

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   time.Second * time.Duration(toInt(c.TimeoutSeconds)),
			KeepAlive: 180 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   2 * time.Second,
		ExpectContinueTimeout: time.Duration(toInt(c.TimeoutSeconds)) * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}

	coll := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(2),
	)

	err := coll.Limit(&colly.LimitRule{DomainGlob: "*", RandomDelay: 1 * time.Second, Parallelism: 6})
	if err != nil {
		return nil, err
	}

	coll.UserAgent = c.UserAgent
	coll.WithTransport(transport)

	coll.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
	})
	coll.OnResponse(func(r *colly.Response) {
		log.Println("response received", r.StatusCode)
	})
	coll.OnError(func(r *colly.Response, err error) {
		log.Println("error:", r.StatusCode, err)
	})

	return coll, nil
}

func (m *CrawlConfigMap) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.SequenceNode {
		return fmt.Errorf("UnmarshalYAML: expected a sequence, got %v", value.Kind)
	}

	configs := make(CrawlConfigMap, len(*m))
	for _, item := range value.Content {
		config := new(Scraper)
		if err := item.Decode(&config); err != nil {
			log.Errorf("UnmarshalYAML - decode error: %v", err)
			return err
		}

		configs[config.Name] = config
	}

	*m = configs
	return nil
}
