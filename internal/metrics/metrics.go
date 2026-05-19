package metrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"proxy/internal/usecase"
	"proxy/pkg/logger"
	"time"
)

var (
	TopUsers = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "top_users_score",
			Help: "Top users by activity request",
		},
		[]string{"rank", "ip", "countRequestDay"},
	)
	ActiveIpsWhite = promauto.NewGauge(prometheus.GaugeOpts{Name: "active_ips_white", Help: "Текущее количество IPS в белом списке"})
	ActiveIpsBlack = promauto.NewGauge(prometheus.GaugeOpts{Name: "active_ips_black", Help: "Текущее количество IPS в черном списке"})
	ActiveIpsGray  = promauto.NewGauge(prometheus.GaugeOpts{Name: "active_ips_gray", Help: "Текущее количество IPS в сером списке"})
)

type MetricsPuller interface {
	Init()
	StartTopUsersUpdater()
	LoadIpsLists()
	Stop()
}

type MetricsPullerImpl struct {
	ListUseCase   usecase.ListUserCase
	RateLimiterUC usecase.RateLimiterUseCase
	ticker        *time.Ticker
	done          chan struct{}
}

func CreateMetricsPuller(listUseCase usecase.ListUserCase, rateLimiterUC usecase.RateLimiterUseCase) MetricsPuller {
	return &MetricsPullerImpl{
		ListUseCase:   listUseCase,
		RateLimiterUC: rateLimiterUC,
		done:          make(chan struct{}),
	}
}

func (mp *MetricsPullerImpl) Init() {
	mp.LoadIpsLists()
	go mp.StartTopUsersUpdater()
}

func (mp *MetricsPullerImpl) LoadIpsLists() {
	whiteList, _ := mp.ListUseCase.GetAllIps("white")
	ActiveIpsWhite.Set(float64(len(whiteList)))

	blackList, _ := mp.ListUseCase.GetAllIps("black")
	ActiveIpsBlack.Set(float64(len(blackList)))

	grayList, _ := mp.ListUseCase.GetAllIps("gray")
	ActiveIpsGray.Set(float64(len(grayList)))
}

func (mp *MetricsPullerImpl) StartTopUsersUpdater() {
	mp.ticker = time.NewTicker(15 * time.Second)
	defer mp.ticker.Stop()

	for {
		select {
		case <-mp.ticker.C:
			mp.updateTopUsers()
		case <-mp.done:
			return
		}
	}
}

func (mp *MetricsPullerImpl) updateTopUsers() {
	logger.Info("Start Update Top User")
	keys, sortedMap := mp.RateLimiterUC.GetIpsByRequest()

	TopUsers.Reset()
	for i := 0; i < 10 && i < len(keys); i++ {
		key := keys[i]
		count := sortedMap[key]
		TopUsers.WithLabelValues(
			fmt.Sprintf("%d", i+1),
			key,
			fmt.Sprintf("%d", count),
		).Set(float64(count))
	}
}

func (mp *MetricsPullerImpl) Stop() {
	if mp.ticker != nil {
		mp.ticker.Stop()
	}
	close(mp.done)
}
