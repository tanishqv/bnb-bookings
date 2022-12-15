package models

import "github.com/tanishqv/bnb-bookings/internal/forms"

// TemplateData holds data sent from the handlers to the templates
type TemplateData struct {
	StringMap map[string]string
	IntMap    map[string]int
	FloatMap  map[string]float32
	Data      map[string]interface{}
	CSRFToken string
	Flash     string
	Warning   string
	Error     string
	Form      *forms.Form
}
