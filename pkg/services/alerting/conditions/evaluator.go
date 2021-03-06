package conditions

import (
	"encoding/json"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/services/alerting"
)

var (
	defaultTypes []string = []string{"gt", "lt"}
	rangedTypes  []string = []string{"within_range", "outside_range"}
)

type AlertEvaluator interface {
	Eval(reducedValue *float64) bool
}

type NoDataEvaluator struct{}

func (e *NoDataEvaluator) Eval(reducedValue *float64) bool {
	return reducedValue == nil
}

type ThresholdEvaluator struct {
	Type      string
	Threshold float64
}

func newThresholdEvaludator(typ string, model *simplejson.Json) (*ThresholdEvaluator, error) {
	params := model.Get("params").MustArray()
	if len(params) == 0 {
		return nil, alerting.ValidationError{Reason: "Evaluator missing threshold parameter"}
	}

	firstParam, ok := params[0].(json.Number)
	if !ok {
		return nil, alerting.ValidationError{Reason: "Evaluator has invalid parameter"}
	}

	defaultEval := &ThresholdEvaluator{Type: typ}
	defaultEval.Threshold, _ = firstParam.Float64()
	return defaultEval, nil
}

func (e *ThresholdEvaluator) Eval(reducedValue *float64) bool {
	switch e.Type {
	case "gt":
		return *reducedValue > e.Threshold
	case "lt":
		return *reducedValue < e.Threshold
	}

	return false
}

type RangedEvaluator struct {
	Type  string
	Lower float64
	Upper float64
}

func newRangedEvaluator(typ string, model *simplejson.Json) (*RangedEvaluator, error) {
	params := model.Get("params").MustArray()
	if len(params) == 0 {
		return nil, alerting.ValidationError{Reason: "Evaluator missing threshold parameter"}
	}

	firstParam, ok := params[0].(json.Number)
	if !ok {
		return nil, alerting.ValidationError{Reason: "Evaluator has invalid parameter"}
	}

	secondParam, ok := params[1].(json.Number)
	if !ok {
		return nil, alerting.ValidationError{Reason: "Evaluator has invalid second parameter"}
	}

	rangedEval := &RangedEvaluator{Type: typ}
	rangedEval.Lower, _ = firstParam.Float64()
	rangedEval.Upper, _ = secondParam.Float64()
	return rangedEval, nil
}

func (e *RangedEvaluator) Eval(reducedValue *float64) bool {
	switch e.Type {
	case "within_range":
		return (e.Lower < *reducedValue && e.Upper > *reducedValue) || (e.Upper < *reducedValue && e.Lower > *reducedValue)
	case "outside_range":
		return (e.Upper < *reducedValue && e.Lower < *reducedValue) || (e.Upper > *reducedValue && e.Lower > *reducedValue)
	}

	return false
}

func NewAlertEvaluator(model *simplejson.Json) (AlertEvaluator, error) {
	typ := model.Get("type").MustString()
	if typ == "" {
		return nil, alerting.ValidationError{Reason: "Evaluator missing type property"}
	}

	if inSlice(typ, defaultTypes) {
		return newThresholdEvaludator(typ, model)
	}

	if inSlice(typ, rangedTypes) {
		return newRangedEvaluator(typ, model)
	}

	if typ == "no_data" {
		return &NoDataEvaluator{}, nil
	}

	return nil, alerting.ValidationError{Reason: "Evaludator invalid evaluator type"}
}

func inSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
