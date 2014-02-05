package eco

import (
	"errors"
)

type TestRecord struct {
	otmcode   string
	diameter  float64
	region    string
	speciesid int
}

type TestingContext struct {
	hasRegions   bool
	singleRegion string
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
	diameter *float64, otmcode *string, speciesid *int, region *string) error {

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
	*region = data.region
	*speciesid = data.speciesid

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

func (t *TestingContext) GetRegionForInstance(instance int) (string, error) {
	// Has region data in the records, return nil
	// indicating that we must grab region info
	if t.hasRegions {
		return "", nil
	}

	return t.singleRegion, nil
}

func (t *TestingContext) RowsForTreesWithRegion(
	where string, params ...string) (Fetchable, error) {
	return Fetchable(t), nil
}

func (t *TestingContext) RowsForTreesWithoutRegion(
	where string, params ...string) (Fetchable, error) {
	return Fetchable(t), nil
}
