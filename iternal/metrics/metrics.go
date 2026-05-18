package metrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
	"proxy/iternal/usecase"
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
	ActiveIpsWhite = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_ips_white",
			Help: "Текущее количество IPS в белом списке",
		},
	)
	ActiveIpsBlack = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_ips_black",
			Help: "Текущее количество IPS в черном списке",
		},
	)
	ActiveIpsGray = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_ips_gray",
			Help: "Текущее количество IPS в сером списке",
		},
	)
)

type MetricsPuller interface {
	Init()
	StartTopUsersUpdater()
	LoadIpsLists()
}

type MetricsPullerImpl struct {
	ListUseCase   usecase.ListUserCase
	RateLimiterUC *usecase.RateLimiterUseCase
}

func CreateMetricsPuller(listUseCase usecase.ListUserCase, rateLimiterUC *usecase.RateLimiterUseCase) MetricsPuller {
	return &MetricsPullerImpl{
		ListUseCase:   listUseCase,
		RateLimiterUC: rateLimiterUC,
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
	ActiveIpsBlack.Set(float64(len(grayList)))
}

func (mp *MetricsPullerImpl) StartTopUsersUpdater() {
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		for range ticker.C {
			logger.Info("Start Update Top User")
			keys, sortedMap := mp.RateLimiterUC.GetIpsByRequst()

			TopUsers.Reset()
			for i := 10; i >= 0; i-- {
				if i >= len(keys) {
					continue
				}
				key := keys[i]
				count := sortedMap[key]
				logger.Info("Top User",
					zap.String("Rank", fmt.Sprintf("%d", i+1)),
					zap.String("User", key))
				TopUsers.WithLabelValues(
					fmt.Sprintf("%d", i+1),
					key,
					fmt.Sprintf("%d", count),
				).Set(float64(count))
			}
		}
	}()
}
