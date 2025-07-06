package csgo

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

var (
	defaultMinFloat = decimal.RequireFromString("0.06")
	defaultMaxFloat = decimal.RequireFromString("0.8")
)

// Paintkit represents the image details of a skin, i.e. the available float
// range the skin can be in. Every entities.Skin has an associated Paintkit.
type Paintkit struct {
	Id          string          `json:"id"`
	Index       int             `json:"index"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	RarityId    string          `json:"rarityId"`
	MinFloat    decimal.Decimal `json:"minFloat"`
	MaxFloat    decimal.Decimal `json:"maxFloat"`
}

// mapToPaintkit converts the provided map into a Paintkit providing
// all required parameters are present and of the correct type.
func mapToPaintkit(data map[string]interface{}, language *language) (*Paintkit, error) {

	response := &Paintkit{
		RarityId: "common", // common is "default" rarity
		MinFloat: defaultMinFloat,
		MaxFloat: defaultMaxFloat,
	}

	// get Name
	if val, err := crawlToType[string](data, "name"); err != nil {
		return nil, errors.New("Id (name) missing from Paintkit")
	} else {
		response.Id = val
	}

	// get language Name Id
	if val, ok := data["description_tag"].(string); ok {
		name, err := language.lookup(val)
		if err != nil {
			return nil, err
		}

		response.Name = name
	}

	// get language Description Id
	if val, ok := data["description_string"].(string); ok {
		description, err := language.lookup(val)
		if err != nil {
			return nil, err
		}

		response.Description = description
	}

	// get min float
	if val, ok := data["wear_remap_min"].(string); ok {
		if valFloat, err := decimal.NewFromString(val); err == nil {
			response.MinFloat = valFloat
		} else {
			return nil, errors.New("Paintkit has non-float min float value (wear_remap_min)")
		}
	}

	// get max float
	if val, ok := data["wear_remap_max"].(string); ok {
		if valFloat, err := decimal.NewFromString(val); err == nil {
			response.MaxFloat = valFloat
		} else {
			return nil, errors.New("Paintkit has non-float max float value (wear_remap_max)")
		}
	}

	return response, nil
}

// getPaintkits gathers all Paintkits in the provided items data and returns them
// as map[paintkitId]Paintkit.
func (c *csgoItems) getPaintkits() (map[int]*Paintkit, error) {

	response := map[int]*Paintkit{}

	rarities, err := crawlToType[map[string]interface{}](c.items, "paint_kits_rarity")
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to extract paint_kits_rarity: %s", err.Error()))
	}

	kits, err := crawlToType[map[string]interface{}](c.items, "paint_kits")
	if err != nil {
		return nil, errors.Wrap(err, "unable to locate paint_kits in provided items")
	}

	for index, kit := range kits {
		mKit, ok := kit.(map[string]interface{})
		iIndex, err := strconv.Atoi(index)
		if !ok {
			return nil, fmt.Errorf("unexpected Paintkit layout in paint_kits for index (%s)", index)
		}

		if err != nil {
			return nil, fmt.Errorf("Unable to change index to int index (%s)", index)
		}

		converted, err := mapToPaintkit(mKit, c.language)
		if err != nil {
			return nil, err
		}

		if converted.Id == "workshop_default" {
			continue
		}

		if rarity, ok := rarities[converted.Id].(string); ok {
			converted.RarityId = rarity
		}

		// if default paintkit, manually set rarity
		if converted.Id == "default" {
			converted.RarityId = "default"
		}

		converted.Index = iIndex
		response[converted.Index] = converted
	}

	return response, nil
}
