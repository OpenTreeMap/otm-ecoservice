package eco

import (
	"errors"
)

type TestRecord struct {
	otmcode   string
	diameter  float64
	x         float64
	y         float64
	speciesid int
}

type regioninfo struct {
	region string
	xcoord float64
}

type TestingContext struct {
	hasRegions   bool
	singleRegion regioninfo
	activeIndex  int

	data []*TestRecord
}

func (t *TestingContext) Reset() {
	t.activeIndex = -1
}

func (t *TestingContext) GetOverrideMap() (map[int]map[string]map[int]string, error) {
	return nil, errors.New("not implemented")
}

func (t *TestingContext) GetDataWithRegion(
	diameter *float64, otmcode *string,
	speciesid *int, x *float64, y *float64) error {

	// Can't call this method
	if !t.hasRegions {
		return errors.New("No regions in this set")
	}

	if t.activeIndex >= len(t.data) {
		panic("Past end of dataset")
	}
	data := t.data[t.activeIndex]

	*diameter = data.diameter
	*otmcode = data.otmcode
	*speciesid = data.speciesid
	*x = data.x
	*y = data.y

	return nil
}

func (t *TestingContext) GetDataWithoutRegion(
	diameter *float64, otmcode *string, speciesid *int) error {

	// Can't call this method
	if t.hasRegions {
		return errors.New("There are regions in this set")
	}

	if t.activeIndex >= len(t.data) {
		panic("Past end of dataset")
	}

	data := t.data[t.activeIndex]

	*diameter = data.diameter
	*otmcode = data.otmcode
	*speciesid = data.speciesid

	return nil
}

func (t *TestingContext) Close() error {
	return nil
}

func (t *TestingContext) Next() bool {
	t.activeIndex += 1

	if t.activeIndex >= len(t.data) {
		return false
	}

	return true
}

func (t *TestingContext) GetRegionsForInstance(
	regions map[int]Region, instance int) ([]Region, error) {

	regionsslice := make([]Region, len(regions))

	for _, v := range regions {
		regionsslice = append(regionsslice, v)
	}

	return regionsslice, nil
}

func (t *TestingContext) ExecSql() (Fetchable, error) {
	return Fetchable(t), nil
}
