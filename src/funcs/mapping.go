package funcs

import (
	"encoding/json"
	"fmt"
	"github.com/cihub/seelog"
	_ "github.com/mattn/go-sqlite3"
	"github.com/smartping/smartping/src/g"
	"github.com/smartping/smartping/src/nettools"
	"net"
	"strconv"
	"sync"
	"time"
)

var (
	MapLock         = new(sync.Mutex)
	MapStatus       map[string][]g.MapVal           // telecom (ctcc/cucc/cmcc) -> []MapVal with Name=province
	MapDetailStatus map[string]g.ProvinceDetail     // province -> detail
	cityDelays      map[string]map[string]map[string]float64 // prov -> tel -> city -> delay
)

func Mapping() {
	var wg sync.WaitGroup
	MapStatus = make(map[string][]g.MapVal)
	MapDetailStatus = make(map[string]g.ProvinceDetail)
	cityDelays = make(map[string]map[string]map[string]float64)

	// Initialize MapStatus for each telecom
	MapStatus["ctcc"] = []g.MapVal{}
	MapStatus["cucc"] = []g.MapVal{}
	MapStatus["cmcc"] = []g.MapVal{}

	seelog.Debug("[func:Mapping]", g.Cfg.Chinamap)
	for prov, telecomMap := range g.Cfg.Chinamap {
		for tel, cityMap := range telecomMap {
			for city, ips := range cityMap {
				if len(ips) > 0 {
					go MappingTask(tel, prov, city, ips, &wg)
					wg.Add(1)
				}
			}
		}
	}
	wg.Wait()

	// Calculate province-level averages from city data
	calculateProvinceAverages()

	MapPingStorage()
}

func calculateProvinceAverages() {
	// Group delays by province and telecom
	provTelDelays := make(map[string]map[string][]float64) // prov -> tel -> []delays

	for prov, telecomMap := range cityDelays {
		for tel, cityMap := range telecomMap {
			for _, delay := range cityMap {
				if provTelDelays[prov] == nil {
					provTelDelays[prov] = make(map[string][]float64)
				}
				provTelDelays[prov][tel] = append(provTelDelays[prov][tel], delay)
			}
		}
	}

	// Calculate averages and store in MapStatus (key=telecom, Name=province)
	for prov, telecomMap := range provTelDelays {
		for tel, delays := range telecomMap {
			var total, effCnt float64
			for _, d := range delays {
				if d < 1000 { // ignore timeout values
					total += d
					effCnt++
				}
			}
			avgDelay := 0.0
			if effCnt > 0 {
				avgDelay = total / effCnt
			}
			MapStatus[tel] = append(MapStatus[tel], g.MapVal{
				Name:  prov,
				Value: avgDelay,
			})
		}
	}
}

func MappingTask(tel string, prov string, city string, ips []string, wg *sync.WaitGroup) {
	seelog.Info(fmt.Sprintf("Start MappingTask %s %s %s..", tel, prov, city))
	var ipDetails []g.IPDetail
	var totalDelay float64
	effCnt := 0

	for _, ip := range ips {
		seelog.Debug("[func:StartChinaMapPing]", ip)
		ipaddr, err := net.ResolveIPAddr("ip", ip)
		stat := g.PingSt{}
		stat.MinDelay = -1
		stat.LossPk = 0

		if err == nil {
			for i := 0; i < 3; i++ {
				delay, err := nettools.RunPing(ipaddr, 3*time.Second, 64, i)
				if err == nil {
					stat.AvgDelay = stat.AvgDelay + delay
					if stat.MaxDelay < delay {
						stat.MaxDelay = delay
					}
					if stat.MinDelay == -1 || stat.MinDelay > delay {
						stat.MinDelay = delay
					}
					stat.RevcPk = stat.RevcPk + 1
					seelog.Debug("[func:StartChinaMapPing IcmpPing] ID:", i, " IP:", ip)
				} else {
					seelog.Debug("[func:StartChinaMapPing IcmpPing] ID:", i, " IP:", ip, " | ", err)
					stat.LossPk = stat.LossPk + 1
				}
				stat.SendPk = stat.SendPk + 1
			}
			if stat.RevcPk > 0 {
				stat.AvgDelay = stat.AvgDelay / float64(stat.RevcPk)
				stat.LossPk = int((float64(stat.SendPk-stat.RevcPk) / float64(stat.SendPk)) * 100)
			} else {
				stat.AvgDelay = 2000
				stat.LossPk = 100
			}
		} else {
			stat.AvgDelay = 2000.00
			stat.MinDelay = 2000.00
			stat.MaxDelay = 2000.00
			stat.SendPk = 0
			stat.RevcPk = 0
			stat.LossPk = 100
		}

		ipDetails = append(ipDetails, g.IPDetail{
			IP:    ip,
			Delay: stat.AvgDelay,
			Loss:  stat.LossPk,
		})
		storePingResult(ip, stat)

		if stat.AvgDelay < 1000 {
			totalDelay += stat.AvgDelay
			effCnt++
		}
	}

	avgDelay := 0.0
	if effCnt > 0 {
		avgDelay = totalDelay / float64(effCnt)
	}

	MapLock.Lock()
	// Store city delay for province averaging
	if cityDelays[prov] == nil {
		cityDelays[prov] = make(map[string]map[string]float64)
	}
	if cityDelays[prov][tel] == nil {
		cityDelays[prov][tel] = make(map[string]float64)
	}
	cityDelays[prov][tel][city] = avgDelay

	// Store detailed IP info
	if MapDetailStatus[prov].Province == "" {
		MapDetailStatus[prov] = g.ProvinceDetail{
			Province: prov,
			Ctcc:     make(map[string][]g.IPDetail),
			Cucc:     make(map[string][]g.IPDetail),
			Cmcc:     make(map[string][]g.IPDetail),
		}
	}
	provDetail := MapDetailStatus[prov]
	switch tel {
	case "ctcc":
		provDetail.Ctcc[city] = ipDetails
	case "cucc":
		provDetail.Cucc[city] = ipDetails
	case "cmcc":
		provDetail.Cmcc[city] = ipDetails
	}
	MapDetailStatus[prov] = provDetail
	MapLock.Unlock()
	wg.Done()
	seelog.Info(fmt.Sprintf("Finish MappingTask %s %s %s..", tel, prov, city))
}

func MapPingStorage() {
	seelog.Info("Start MapPingStorage...")
	seelog.Debug(MapStatus)
	jdata, err := json.Marshal(MapStatus)
	if err != nil {
		seelog.Error("[func:StartPing] Json Error ", err)
	}
	sql := "REPLACE INTO [mappinglog] (logtime, mapjson) values('" + time.Now().Format("2006-01-02 15:04") + "','" + string(jdata) + "')"
	g.DLock.Lock()
	g.Db.Exec(sql)
	_, err = g.Db.Exec(sql)
	seelog.Debug(sql)
	if err != nil {
		seelog.Error("[func:StartPing] Sql Error ", err)
	}
	g.DLock.Unlock()
	seelog.Debug("[func:MapPingStorage] ", sql)
	seelog.Info("Finish MapPingStorage...")
}

func storePingResult(ip string, stat g.PingSt) {
	logtime := time.Now().Format("2006-01-02 15:04")
	sql := "INSERT INTO [pinglog] (logtime, target, maxdelay, mindelay, avgdelay, sendpk, revcpk, losspk) values('" + logtime + "','" + ip + "','" + strconv.FormatFloat(stat.MaxDelay, 'f', 2, 64) + "','" + strconv.FormatFloat(stat.MinDelay, 'f', 2, 64) + "','" + strconv.FormatFloat(stat.AvgDelay, 'f', 2, 64) + "','" + strconv.Itoa(stat.SendPk) + "','" + strconv.Itoa(stat.RevcPk) + "','" + strconv.Itoa(stat.LossPk) + "')"
	_, err := g.Db.Exec(sql)
	if err != nil {
		seelog.Error("[func:storePingResult] Sql Error ", err)
	}
}
