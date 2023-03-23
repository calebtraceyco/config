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

type Scraper struct {
	Name                  yaml.Node `yaml:"Name"`
	AppsJSONPath          yaml.Node `yaml:"AppsJSONPath"`
	TimeoutSeconds        yaml.Node `yaml:"TimeoutSeconds"`
	LoadingTimeoutSeconds yaml.Node `yaml:"LoadingTimeoutSeconds"`
	JSON                  yaml.Node `yaml:"JSON"`
	MaxDepth              yaml.Node `yaml:"MaxDepth"`
	//visitedLinks           yaml.Node `yaml:"VisitedLinks"`
	MaxVisitedLinks        yaml.Node `yaml:"MaxVisitedLinks"`
	MsDelayBetweenRequests yaml.Node `yaml:"MsDelayBetweenRequests"`
	UserAgent              yaml.Node `yaml:"UserAgent"`
	Collector              *colly.Collector
	//componentConfigs       ComponentConfigs
}

type CrawlConfigMap map[string]*Scraper

func (c *Scraper) collector() (*colly.Collector, error) {

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   time.Second * time.Duration(toInt(c.TimeoutSeconds.Value)),
			KeepAlive: 180 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   2 * time.Second,
		ExpectContinueTimeout: time.Duration(toInt(c.TimeoutSeconds.Value)) * time.Second,
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

	coll.UserAgent = c.UserAgent.Value
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

func (cm *CrawlConfigMap) UnmarshalYAML(node *yaml.Node) error {
	*cm = CrawlConfigMap{}
	var crawlers []Scraper

	if decodeErr := node.Decode(&crawlers); decodeErr != nil {
		return fmt.Errorf("decode error: %v", decodeErr.Error())
	}

	for _, crawler := range crawlers {
		var crawlerKey string
		crawlerCopy := crawler

		if serviceErr := crawler.Name.Decode(&crawlerKey); serviceErr != nil {
			return fmt.Errorf("decode error: %w", serviceErr)
		}
		(*cm)[crawlerKey] = &crawlerCopy
	}
	return nil
}
