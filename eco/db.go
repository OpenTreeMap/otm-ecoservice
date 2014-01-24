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
	diameter *float64, otmcode *string, region *string) error {
	return (*sql.Rows)(dbr).Scan(diameter, otmcode, region)
}

func (dbr *DBRow) GetDataWithoutRegion(
	diameter *float64, otmcode *string) error {
	return (*sql.Rows)(dbr).Scan(diameter, otmcode)
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

func (dbc *DBContext) GetRegionForInstance(instance int) (string, error) {
	db := (*sql.DB)(dbc)
	rows, err :=
		db.Query(`select itree_region_default
                          from treemap_instance
                          where treemap_instance.id = $1`,
			instance)

	if err != nil {
		return "", err
	}

	defer rows.Close()

	region := ""

	rows.Next()
	rows.Scan(&region)

	// The instance has specified a forced override
	// so bail now, returning that value
	if len(region) > 0 {
		return region, nil
	}

	// We can save a ton of time by pre-loading
	// region data if a given instance only
	// intersects a single eco region
	rows, err =
		db.Query(`select distinct code
                          from treemap_instance
                             left join treemap_itreeregion
                             on st_intersects(
                                treemap_itreeregion.geometry,
                                treemap_instance.bounds)
                          where
                             code is not null and
                          treemap_instance.id = $1 limit 2`, instance)

	if err != nil {
		return "", err
	}

	defer rows.Close()

	if rows.Next() {
		rows.Scan(&region)

		// If true there are multiple rows in this
		// query so we can't take this shortcut
		if rows.Next() {
			return "", nil
		} else {
			// Only one, and we found it!
			return region, nil
		}
	}

	// The query didn't return anything at all
	// which means that we wont have any data, but
	// that will be handled at a future step
	return "", nil
}

func convertInterface(params []string) []interface{} {
	paramsi := make([]interface{}, len(params))

	for i, v := range params {
		paramsi[i] = interface{}(v)
	}

	return paramsi
}

func (dbc *DBContext) RowsForTreesWithoutRegion(
	where string, params ...string) (Fetchable, error) {
	db := (*sql.DB)(dbc)

	query := fmt.Sprintf(
		`select diameter, treemap_species.otm_code
                    from treemap_species, treemap_tree
                    where
                       %v and
                       treemap_species.id = treemap_tree.species_id and
                       diameter is not null`, where)

	paramsi := convertInterface(params)

	rows, err := db.Query(query, paramsi...)

	return (*DBRow)(rows), err
}

func (dbc *DBContext) RowsForTreesWithRegion(
	where string, params ...string) (Fetchable, error) {
	db := (*sql.DB)(dbc)

	query := fmt.Sprintf(
		`select
                    diameter, treemap_species.otm_code,
                    treemap_itreeregion.code
                 from
                    treemap_species,
                    treemap_tree,
                    treemap_mapfeature
                    left join treemap_itreeregion
                       on ST_Contains(
                          treemap_itreeregion.geometry,
                          treemap_mapfeature.the_geom_webmercator)
                 where
                    %v and
                    treemap_mapfeature.id = treemap_tree.plot_id and
                    treemap_species.id = treemap_tree.species_id and
                    treemap_itreeregion.code is not null and
                    diameter is not null and
                    otm_code is not null and
                    treemap_tree.instance_id = $1
                 `, where)

	paramsi := convertInterface(params)

	rows, err := db.Query(query, paramsi...)

	return (*DBRow)(rows), err
}
