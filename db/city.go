package db

import (
	"encoding/csv"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	citystore "hr-tools-backend/lib/dicts/city/store"
	dbmodels "hr-tools-backend/models/db"
	"os"
	"strconv"
)

func fillCities() {
	log.Info("предзаполнение городов")
	cityStore := citystore.NewInstance(DB)
	scopeList, err := cityStore.List("")
	if err != nil {
		log.WithError(err).Error("ошибка предзаполнения городов")
		return
	}
	if len(scopeList) > 0 {
		log.Info("города заполнены")
		return
	}

	lines, err := readCsvFile("./static_preload/city.csv", ';')
	if err != nil {
		log.WithError(err).Error("ошибка загрузки файла с городами")
		return
	}
	for k, line := range lines {
		lat, err := strconv.ParseFloat(line[20], 64)
		if err != nil {
			log.WithError(err).Errorf("ошибка загрузки файла с городами, строка %v", k)
			return
		}
		lon, err := strconv.ParseFloat(line[21], 64)
		if err != nil {
			log.WithError(err).Errorf("ошибка загрузки файла с городами, строка %v", k)
			return
		}
		rec := dbmodels.City{
			BaseModel: dbmodels.BaseModel{
				ID: line[13],
			},
			Address:         line[0],
			PostalCode:      line[1],
			Country:         line[2],
			FederalDistrict: line[3],
			RegionType:      line[4],
			Region:          line[5],
			AreaType:        line[6],
			Area:            line[7],
			CityType:        line[8],
			City:            line[9],
			SettlementType:  line[10],
			Settlement:      line[11],
			Okato:           line[16],
			Oktmo:           line[17],
			Timezone:        line[19],
			Lat:             lat,
			Lon:             lon,
		}
		err = cityStore.Add(rec, true)
		if err != nil {
			log.
				WithError(err).
				WithField("address", rec.Address).
				WithField("code", rec.ID).
				Error("ошибка добавления городов")
			return
		}
	}

	log.Info("города добавлены")
}

func readCsvFile(filePath string, comma rune) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка открытия файла")
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	csvReader.Comma = comma
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, errors.Wrap(err, "ошибка обработки файла")
	}

	return records, nil
}
