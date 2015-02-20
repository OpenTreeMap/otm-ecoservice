package eco

import (
	"fmt"
	"math/rand"
	"testing"
)

// Data sanity checks
func TestRegionsArePresent(t *testing.T) {
	// Spot check NoEastXXX and LoMidWXXX
	m, _ := LoadSpeciesMap("../data/species.json")

	_, contained := m["NoEastXXX"]

	if !contained {
		t.Fatal("Missing NoEastXXX in master species table")
	}

	_, contained = m["LoMidWXXX"]

	if !contained {
		t.Fatal("Missing NoEastXXX in master species table")
	}

	// Also check in base dictionary
	l := LoadFiles("../data/")

	_, contained = l["NoEastXXX"]

	if !contained {
		t.Fatal("Missing NoEastXXX in master species table")
	}

	_, contained = l["LoMidWXXX"]

	if !contained {
		t.Fatal("Missing NoEastXXX in master species table")
	}
}

func TestDataMappings(t *testing.T) {
	otmcode := "MASO"

	expected := map[string]string{
		"PiedmtCLT": "BDS OTHER",
		"NoEastXXX": "BDS OTHER",
		"CaNCCoJBK": "BDS OTHER",
		"InlValMOD": "MAGR",
		"SoCalCSMA": "BDS OTHER",
		"GulfCoCHS": "BDS OTHER",
		"CenFlaXXX": "BDS OTHER",
		"PacfNWLOG": "BDS OTHER",
		"InlEmpCLM": "MAGR",
	}

	m, _ := LoadSpeciesMap("../data/species.json")

	for region, target := range expected {
		regionmap, found := m[region]

		if !found {
			t.Fatalf("Missing region %v", region)
		}

		itreecode, found := regionmap[otmcode]

		if !found {
			t.Fatalf("Missing data for code %v in region %v",
				itreecode, region)
		}

		if itreecode != target {
			t.Fatalf("Invalid itree code for otmcode %v "+
				"in region %v (got %v, expected %v)",
				itreecode, region, itreecode, target)
		}
	}
}

func TestGetITreeCodesByRegion(t *testing.T) {
	regionData := LoadFiles("../data/")
	codesByRegion := GetITreeCodesByRegion(regionData)

	region := "NoEastXXX"
	myCode := "ACPL"

	codes, found := codesByRegion[region]
	if !found {
		t.Fatalf("Missing region %v", region)
	}

	for _, code := range codes {
		if code == myCode {
			return
		}
	}
	t.Fatalf("Missing code %v", myCode)
}

func TestSimpleInter(t *testing.T) {
	breaks := []float64{1.0, 3.0}
	values := []float64{4.0, 6.0}

	itreecode := "blah"

	datafile := &Datafile{breaks, map[string][]float64{itreecode: values}}
	datafiles := []*Datafile{datafile}

	result := []float64{0.0}

	CalcOneTree(
		datafiles,
		itreecode,
		2.0,
		result)

	if result[0] != 5.0 {
		t.Fatalf("Expected %v, got %v", 2.0, result[0])
	}
}

// Since there isn't really a canonical benefits
// library to test against, we're just going to
// use a couple of exsiting OTM instances
//
// These test are pretty data dependent but the
// whole purpose of the library is to provide access to
// that data
func TestSpecificTreeData(t *testing.T) {
	region := "LoMidWXXX"
	otmcode := "ULAM"
	dbh := 11.0

	targets := map[string]float64{
		"aq_nox_avoided":     0.01548490,
		"aq_nox_dep":         0.00771784,
		"aq_pm10_avoided":    0.00546863,
		"aq_pm10_dep":        0.016322,
		"aq_sox_avoided":     0.06590,
		"aq_sox_dep":         0.0057742,
		"aq_voc_avoided":     0.0054686,
		"bvoc":               0,
		"co2_avoided":        12.0864829,
		"co2_sequestered":    51.42926,
		"co2_storage":        110.79107,
		"electricity":        12.180839,
		"hydro_interception": 2.5919028,
		"natural_gas":        -18.345013}

	l := LoadFiles("../data/")
	m, _ := LoadSpeciesMap("../data/species.json")

	itreecode := m[region][otmcode]
	factorDataForRegion := l[region]

	factorsum := make([]float64, len(Factors))

	CalcOneTree(
		factorDataForRegion,
		itreecode,
		dbh,
		factorsum)

	factormap := FactorArrayToMap(factorsum)

	for factor, target := range targets {
		calcd := factormap[factor]

		if int(calcd*100000) != int(target*100000) {
			t.Fatalf("Expected %v, got %v for factor %v",
				target, calcd, factor)
		}
	}
}

func generateSpeciesListFromRegion(
	speciesdata map[string]map[string]string,
	targetLength int, region Region) []*TestRecord {

	speciesmap := speciesdata[region.Code]

	possibleSpecies := make([]string, len(speciesmap))
	i := 0
	for k, _ := range speciesmap {
		possibleSpecies[i] = k
		i++
	}

	idx := rand.Perm(targetLength)

	data := make([]*TestRecord, targetLength)

	i = 0
	for sidx := range idx {
		x, y := GetXYOnSurface(region.geom)
		otmcode := possibleSpecies[sidx%len(possibleSpecies)]
		diameter := rand.Float64() * 100.0
		data[i] = &TestRecord{otmcode, diameter, x, y, sidx}
		i++
	}

	return data
}

func benchmarkTreesSingleRegion(targetLength int, b *testing.B) {
	info := regionInfos[0]

	benchmarkTreesMultiRegion([]regioninfo{info},
		targetLength, b)
}

func benchmarkTreesMultiRegion(
	regions []regioninfo, targetLength int, b *testing.B) {

	benchmarkTreesMultiRegionWithOverrides(nil, regions, targetLength, b)
}

func makeSurface(x float64) Geom {
	x1, y1 := x, 0.0
	x2, y2 := x1+1.0, 3.0

	shapewkt := fmt.Sprintf(
		"POLYGON((%f %f, %f %f, %f %f, %f %f, %f %f))",
		x1, y1,
		x1, y2,
		x2, y2,
		x2, y1,
		x1, y1)

	return MakeGeosGeom(shapewkt)
}

func benchmarkTreesMultiRegionWithOverrides(
	overrides map[string]map[int]string,
	regioninfos []regioninfo,
	targetLength int, b *testing.B) {

	region := ""
	if len(regioninfos) == 1 {
		region = regioninfos[0].region
	}

	regions := make([]Region, len(regioninfos))

	InitGeos()

	for i, v := range regioninfos {
		regions[i] = Region{v.region, makeSurface(v.xcoord)}
	}

	l := LoadFiles("../data/")
	speciesdata, _ := LoadSpeciesMap("../data/species.json")

	targetLengthPerRegion := targetLength / len(regions)
	data := make([]*TestRecord, 0)

	for i := range regions {
		newdata := generateSpeciesListFromRegion(
			speciesdata, targetLengthPerRegion,
			regions[i])
		data = append(data, newdata...)
	}

	testingContext := &TestingContext{
		len(regions) > 1, regioninfos[0], 0, data}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testingContext.Reset()
		data, err := CalcBenefitsWithData(
			regions, testingContext, region, speciesdata,
			l, overrides)

		if err != nil {
			b.Fatalf("error: %v", err)
		}
		benchdump = data
	}

	for _, r := range regions {
		GeosDestroy(r.geom)
	}

}

func regionsToRegionInfo(r []string) []regioninfo {
	infos := make([]regioninfo, len(r))

	for i, v := range r {
		infos[i] = regioninfo{v, geomCorners[v]}
	}

	return infos
}

var (
	// Store stuff to this variable to prevent the compiler
	// from getting to tricky
	benchdump interface{}

	regions []string = []string{"PiedmtCLT", "NoEastXXX", "CaNCCoJBK",
		"InlValMOD", "SoCalCSMA", "GulfCoCHS",
		"CenFlaXXX", "PacfNWLOG", "InlEmpCLM"}

	geomCorners = map[string]float64{
		"PiedmtCLT": 18.0,
		"NoEastXXX": 2.0,
		"CaNCCoJBK": 4.0,
		"InlValMOD": 6.0,
		"SoCalCSMA": 8.0,
		"GulfCoCHS": 10.0,
		"CenFlaXXX": 12.0,
		"PacfNWLOG": 14.0,
		"InlEmpCLM": 16.0,
	}

	regionInfos = regionsToRegionInfo(regions)
)

func BenchmarkTreesMultiRegion100(b *testing.B) {
	benchmarkTreesMultiRegion(regionInfos, 1e2, b)
}

func BenchmarkTreesMultiRegion1k(b *testing.B) {
	benchmarkTreesMultiRegion(regionInfos, 1e3, b)
}

func BenchmarkTreesMultiRegion10k(b *testing.B) {
	benchmarkTreesMultiRegion(regionInfos, 1e4, b)
}

func BenchmarkTreesMultiRegion100k(b *testing.B) {
	benchmarkTreesMultiRegion(regionInfos, 1e5, b)
}

func BenchmarkTreesMultiRegion1M(b *testing.B) {
	benchmarkTreesMultiRegion(regionInfos, 1e6, b)
}

func BenchmarkTreesSingleRegion100(b *testing.B)  { benchmarkTreesSingleRegion(1e2, b) }
func BenchmarkTreesSingleRegion1k(b *testing.B)   { benchmarkTreesSingleRegion(1e3, b) }
func BenchmarkTreesSingleRegion10k(b *testing.B)  { benchmarkTreesSingleRegion(1e4, b) }
func BenchmarkTreesSingleRegion100k(b *testing.B) { benchmarkTreesSingleRegion(1e5, b) }
func BenchmarkTreesSingleRegion1M(b *testing.B)   { benchmarkTreesSingleRegion(1e6, b) }
