package masaimodels

import (
	"encoding/base64"
	"strings"

	"github.com/pkg/errors"
)

type GradioResponse struct {
	Elements []GradioUpdate
}

func (g GradioResponse) GetRecognizedText() string {
	for _, elem := range g.Elements {
		if elem.Info == "Recognized text" {
			value, _ := elem.ToString()
			return value
		}
	}
	return ""
}

type GradioUpdate struct {
	Info      string   `json:"info,omitempty"`
	Container bool     `json:"container,omitempty"`
	Visible   bool     `json:"visible,omitempty"`
	ElemClass []string `json:"elem_classes,omitempty"`
	Value     any      `json:"value"`
	Type      string   `json:"__type__"`
}

type PlotValue struct {
	Type string `json:"type"`
	Plot string `json:"plot"`
}

func (g GradioUpdate) IsPlotValue() bool {
	_, ok := g.Value.(PlotValue)
	return ok
}

func (g GradioUpdate) IsStringValue() bool {
	_, ok := g.Value.(string)
	return ok
}

func (g GradioUpdate) ToString() (string, bool) {
	value, ok := g.Value.(string)
	return value, ok
}

func (g GradioUpdate) ToPlotValue() (PlotValue, bool) {
	value, ok := g.Value.(PlotValue)
	return value, ok
}

func (p PlotValue) ToByteArr() (contentType string, body []byte, err error) {
	data := strings.Split(p.Plot, ";")
	if len(data) != 2 {
		return "", nil, errors.New("некорректный формат")
	}
	contentType = data[0]
	bodyBase := data[1]
	bodyBase = strings.Replace(bodyBase, "base64,", "", 1)
	body, err = base64.StdEncoding.DecodeString(bodyBase)
	if err != nil {
		return "", nil, err
	}
	return contentType, body, nil
}
