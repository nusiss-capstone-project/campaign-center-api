package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// LandingPageRepeatableItem is one steps/faq entry stored in JSON columns.
type LandingPageRepeatableItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// LandingPageRepeatableItems is a JSON array of repeatable landing-page items.
type LandingPageRepeatableItems []LandingPageRepeatableItem

// Value implements driver.Valuer.
func (items LandingPageRepeatableItems) Value() (driver.Value, error) {
	if items == nil {
		items = LandingPageRepeatableItems{}
	}
	b, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan implements sql.Scanner.
func (items *LandingPageRepeatableItems) Scan(value any) error {
	if items == nil {
		return fmt.Errorf("LandingPageRepeatableItems: Scan on nil pointer")
	}
	if value == nil {
		*items = LandingPageRepeatableItems{}
		return nil
	}
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("LandingPageRepeatableItems: unsupported Scan type %T", value)
	}
	if len(raw) == 0 {
		*items = LandingPageRepeatableItems{}
		return nil
	}
	var out LandingPageRepeatableItems
	if err := json.Unmarshal(raw, &out); err != nil {
		return err
	}
	if out == nil {
		out = LandingPageRepeatableItems{}
	}
	*items = out
	return nil
}
