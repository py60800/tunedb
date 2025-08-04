module github.com/py60800/tunedb/internal/zdb

go 1.23.4

require (
	github.com/dhowden/tag v0.0.0-20240417053706-3d75831295e8
	github.com/mozillazg/go-unidecode v0.2.0
	github.com/py60800/tunedb/internal/search v0.0.0-00010101000000-000000000000
	github.com/py60800/tunedb/internal/svgtab v0.0.0-00010101000000-000000000000
	github.com/py60800/tunedb/internal/util v0.0.0
	github.com/py60800/tunedb/internal/zixml v0.0.0-20250804120328-3f23783e0f5f
	gorm.io/driver/sqlite v1.5.7
	gorm.io/gorm v1.25.12
)

require (
	github.com/antchfx/xpath v0.0.0-20170515025933-1f3266e77307 // indirect
	github.com/gotk3/gotk3 v0.6.3 // indirect
	github.com/hookttg/svgparser v1.1.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/subchen/go-xmldom v1.1.2 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/api v0.243.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/py60800/tunedb/internal/util => ../util

replace github.com/py60800/tunedb/internal/search => ../search

replace github.com/py60800/tunedb/internal/svgtab => ../svgtab
