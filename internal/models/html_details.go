package models

import (
	"context"
	"net/http"
	"time"
)

type HTMLVersionNumber string

const (
	HTMLVersion5   HTMLVersionNumber = "5"
	HTMLVersion401 HTMLVersionNumber = "4.01"
)

type Heading string

const (
	H1 Heading = "h1"
	H2 Heading = "h2"
	H3 Heading = "h3"
	H4 Heading = "h4"
	H5 Heading = "h5"
	H6 Heading = "h6"
)

type LinkType string

const (
	LinkTypeInternal LinkType = "internal"
	LinkTypeExternal LinkType = "external"
)

type Links map[string]*Link

type HTMLDetails struct {
	Version         *HTMLVersion
	Title           string
	HeadingsCounter map[Heading]int
	Links           Links
	HasLoginForm    bool
}

type HTMLVersion struct {
	Number HTMLVersionNumber
	Strict bool
}

type Link struct {
	URL        string
	Count      int
	Type       LinkType
	Accessible bool
}

func (l Links) AddInternalLink(url string) Links {
	return l.addLink(url, LinkTypeInternal)
}

func (l Links) AddExternalLink(url string) Links {
	return l.addLink(url, LinkTypeExternal)
}

func (l Links) addLink(url string, linkType LinkType) Links {
	if l == nil {
		l = make(map[string]*Link)
	}

	if _, ok := l[url]; ok {
		l[url].Count++
		return l
	}
	link := &Link{
		URL:   url,
		Count: 1,
		Type:  linkType,
	}
	l[url] = link
	return l
}

func (l *Link) VerifyLink() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, l.URL, nil)
	if err != nil {
		l.Accessible = false
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		l.Accessible = false
		return
	}
	if resp.StatusCode != http.StatusOK {
		l.Accessible = false
		return
	}
	l.Accessible = true
}

func (l Links) CountInternalLinks() int {
	return l.countLinksByType(LinkTypeInternal)
}

func (l Links) CountInternalLinksAccessible() int {
	return l.countLinksByTypeAndAccessible(LinkTypeInternal, true)
}

func (l Links) CountInternalLinksInaccessible() int {
	return l.countLinksByTypeAndAccessible(LinkTypeInternal, false)
}

func (l Links) CountExternalLinksAccessible() int {
	return l.countLinksByTypeAndAccessible(LinkTypeExternal, true)
}

func (l Links) CountExternalLinksInaccessible() int {
	return l.countLinksByTypeAndAccessible(LinkTypeExternal, false)
}

func (l Links) countLinksByTypeAndAccessible(linkType LinkType, isAccessible bool) int {
	var counter int
	for _, v := range l {
		if v.Type == linkType && v.Accessible == isAccessible {
			counter += v.Count
		}
	}
	return counter
}

func (l Links) CountExternalLinks() int {
	return l.countLinksByType(LinkTypeExternal)
}

func (l Links) countLinksByType(linkType LinkType) int {
	var counter int
	for _, v := range l {
		if v.Type == linkType {
			counter += v.Count
		}
	}
	return counter
}

func (l Links) GetInternalLinks() []Link {
	return l.getLinksByType(LinkTypeInternal)
}

func (l Links) GetExternalLinks() []Link {
	return l.getLinksByType(LinkTypeExternal)
}

func (l Links) getLinksByType(linkType LinkType) []Link {
	var links []Link
	for _, v := range l {
		if v.Type == linkType {
			links = append(links, *v)
		}
	}
	return links
}
