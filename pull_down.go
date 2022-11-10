package excelizex

import (
	"errors"
	"strings"
)

type pullDown struct {
	find       map[string]*[]pullDown
	nextDimKey string
	value      string
	next       []pullDown
}

func NewPullDown() *pullDown {
	return &pullDown{
		find: make(map[string]*[]pullDown),
	}
}

func (pd *pullDown) Add(data []string, linkTarget ...string) *pullDown {
	var path string
	if len(linkTarget) > 0 {
		path = strings.Join(linkTarget, "&")
	}

	dim := 1
	if path != "" {
		dim = dim + len(linkTarget)
	}

	var pls []pullDown
	for _, d := range data {
		pl := pullDown{
			value: d,
			next:  make([]pullDown, 0),
		}

		if dim == 1 {
			pl.nextDimKey = path
		} else {
			pl.nextDimKey = path + "&" + d
		}

		pls = append(pls, pl)

		if dim == 1 {
			pd.find[d] = &pl.next
		} else {
			pd.find[path+"&"+d] = &pl.next
		}
	}

	var (
		ok   bool
		next *[]pullDown
	)
	if next, ok = pd.find[path]; !ok {
		if dim > 1 {
			panic(errors.New(" cannot find link path " + path))
		}

		pd.next = pls
	} else {
		*next = pls
	}

	return pd
}

func (pd *pullDown) getDimension(dimData *[]map[string][]string) {
	if pd.next == nil {
		return
	}

	for _, s := range pd.next {
		dim := len(strings.Split(s.nextDimKey, "&"))
		switch dim {
		case 1:
			if (*dimData)[0] == nil {
				(*dimData)[0] = make(map[string][]string)
			}

			(*dimData)[0][s.value] = append((*dimData)[0][s.value], s.value)
		case 2:
			if (*dimData)[1] == nil {
				(*dimData)[1] = make(map[string][]string)
			}

			(*dimData)[1][s.nextDimKey] = append((*dimData)[0][s.nextDimKey], s.value)
		default:
			pl := s.find[s.nextDimKey]

			if (*dimData)[dim] == nil {
				(*dimData)[dim] = make(map[string][]string)
			}

			for _, value := range *pl {
				(*dimData)[dim][s.nextDimKey] = append((*dimData)[dim][s.nextDimKey], value.value)
				s.getDimension(dimData)
			}
		}

		s.getDimension(dimData)
	}

	return
}

func (pd *pullDown) Set(cols ...string) SheetOption {
	return func(s *Sheet) {
		if len(cols) <= 0 {
			panic(errors.New(""))
		}

		var dimData = make([]map[string][]string, len(cols))
		pd.getDimension(&dimData)

		if len(dimData) > 0 {

		}
	}
}
