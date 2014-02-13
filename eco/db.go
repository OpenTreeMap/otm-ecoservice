package eco

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type DBInfo struct {
	User     string
	Password string
	Host     string
	Database string
}

type DBRow sql.Rows

func (dbr *DBRow) GetDataWithRegion(
	diameter *float64,
	otmcode *string,
	speciesid *int,
	x *float64,
	y *float64) error {

	err := (*sql.Rows)(dbr).Scan(diameter, speciesid, otmcode, x, y)

	if err == nil {
		*diameter *= CentimetersPerInch
	}

	return err
}

func (dbr *DBRow) GetDataWithoutRegion(
	diameter *float64, otmcode *string, speciesid *int) error {

	err := (*sql.Rows)(dbr).Scan(diameter, speciesid, otmcode)

	if err == nil {
		*diameter *= CentimetersPerInch
	}

	return err
}

func (dbr *DBRow) Close() error {
	return (*sql.Rows)(dbr).Close()
}

func (dbr *DBRow) Next() bool {
	return (*sql.Rows)(dbr).Next()
}

type DBContext sql.DB

func OpenDatabaseConnection(info *DBInfo) (*sql.DB, error) {
	cxnString := fmt.Sprintf("user=%v dbname=%v password=%v host=%v",
		info.User, info.Database, info.Password, info.Host)

	return sql.Open("postgres", cxnString)
}

func (dbc *DBContext) GetRegionGeoms() (map[int]Region, error) {
	db := (*sql.DB)(dbc)

	rows, err :=
		db.Query("select id, code, ST_AsText(geometry) from treemap_itreeregion")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	geoms := make(map[int]Region)

	code := ""
	wkt := ""
	id := 0

	for rows.Next() {
		rows.Scan(&id, &code, &wkt)

		geoms[id] = Region{code, MakeGeosGeom(wkt)}
	}

	return geoms, nil
}

func (dbc *DBContext) GetRegionsForInstance(
	fregions map[int]Region, instance int) ([]Region, error) {

	db := (*sql.DB)(dbc)

	rows, err :=
		db.Query(`select treemap_itreeregion.id
                          from treemap_instance
                             left join treemap_itreeregion
                             on st_intersects(
                                treemap_itreeregion.geometry,
                                treemap_instance.bounds)
                          where
                             code is not null and
                          treemap_instance.id = $1`, instance)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	intersectingRegions := make([]Region, 0)

	id := 0

	for rows.Next() {
		rows.Scan(&id)

		intersectingRegions = append(
			intersectingRegions, regions[id])
	}

	return intersectingRegions, nil
}

func (dbc *DBContext) ExecSql(query string) (Fetchable, error) {
	db := (*sql.DB)(dbc)

	fmt.Println(query)

	rows, err := db.Query(query)

	return (*DBRow)(rows), err
}

func (dbc *DBContext) GetOverrideMap() (map[int]map[string]map[int]string, error) {
	db := (*sql.DB)(dbc)

	overrides := make(map[int]map[string]map[int]string)
	query := `select
		    itree_code,
		    treemap_itreeregion.code,
		    treemap_species.id,
		    treemap_species.instance_id
		  from
		    treemap_itreecodeoverride,
		    treemap_itreeregion,
		    treemap_species
		  where
		    treemap_itreecodeoverride.region_id =
		      treemap_itreeregion.id AND
		    treemap_itreecodeoverride.instance_species_id =
		      treemap_species.id
		  `

	rows, err := db.Query(query)

	if err != nil {
		return nil, err
	}

	code, region, sid, iid := "", "", 0, 0

	for rows.Next() {
		rows.Scan(&code, &region, &sid, &iid)

		regionsmap, found := overrides[iid]

		if !found {
			regionsmap = make(map[string]map[int]string)
			overrides[iid] = regionsmap
		}

		sidmap, found := regionsmap[region]

		if !found {
			sidmap = make(map[int]string)
			regionsmap[region] = sidmap
		}

		sidmap[sid] = code
	}

	return overrides, nil
}
