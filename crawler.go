package config_yaml

import (
	"crypto/tls"
	"fmt"
	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"net"
	"net/http"
	"strconv"
	"time"
)

type CrawlerConfig struct {
	Name                   yaml.Node `yaml:"Name"`
	AppsJSONPath           yaml.Node `yaml:"AppsJSONPath"`
	TimeoutSeconds         yaml.Node `yaml:"TimeoutSeconds"`
	LoadingTimeoutSeconds  yaml.Node `yaml:"LoadingTimeoutSeconds"`
	JSON                   yaml.Node `yaml:"JSON"`
	MaxDepth               yaml.Node `yaml:"MaxDepth"`
	visitedLinks           yaml.Node `yaml:"VisitedLinks"`
	MaxVisitedLinks        yaml.Node `yaml:"MaxVisitedLinks"`
	MsDelayBetweenRequests yaml.Node `yaml:"MsDelayBetweenRequests"`
	UserAgent              yaml.Node `yaml:"UserAgent"`
	Collector              *colly.Collector
	componentConfigs       ComponentConfigs
}

type CrawlConfigMap map[string]*CrawlerConfig

func (c *CrawlerConfig) CrawlerComponentConfigs() ComponentConfigs {
	return c.componentConfigs
}

func (c *CrawlerConfig) CrawlerService() (*colly.Collector, error) {
	timeout, _ := strconv.Atoi(c.TimeoutSeconds.Value)

	tp := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   time.Second * time.Duration(timeout),
			KeepAlive: 180 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   2 * time.Second,
		ExpectContinueTimeout: time.Duration(timeout) * time.Second,
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
	coll.WithTransport(tp)

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

func (ccm *CrawlConfigMap) UnmarshalYAML(node *yaml.Node) error {
	*ccm = CrawlConfigMap{}
	var crawlers []CrawlerConfig

	if decodeErr := node.Decode(&crawlers); decodeErr != nil {
		return fmt.Errorf("decode error: %v", decodeErr.Error())
	}

	for _, c := range crawlers {
		var cString string
		cCopy := c
		serviceErr := c.Name.Decode(&cString)
		if serviceErr != nil {
			return fmt.Errorf("decode error: %v", serviceErr.Error())
		}
		(*ccm)[cString] = &cCopy
	}

	return nil
}
