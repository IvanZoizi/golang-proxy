package metrics

import (
	"testing"
	"time"

	"proxy/internal/entity"
)

type mockListUseCaseForMetrics struct {
	white []string
	black []string
	gray  []string
}

func (m *mockListUseCaseForMetrics) GetAllIps(listName string) ([]string, error) {
	switch listName {
	case "white":
		return m.white, nil
	case "black":
		return m.black, nil
	case "gray":
		return m.gray, nil
	}
	return nil, nil
}
func (m *mockListUseCaseForMetrics) AddIp(ip, listName string) error                 { return nil }
func (m *mockListUseCaseForMetrics) Contains(ip, listName string) (bool, error)      { return false, nil }
func (m *mockListUseCaseForMetrics) DeleteIp(ip, listName string) error              { return nil }
func (m *mockListUseCaseForMetrics) GetAllCIDRIps(listName string) ([]string, error) { return nil, nil }
func (m *mockListUseCaseForMetrics) GetAllRangeIps(listName string) ([]string, error) {
	return nil, nil
}

type mockRateLimiterForMetrics struct {
	ips   []string
	count map[string]int
}

func (m *mockRateLimiterForMetrics) GetIpsByRequest() ([]string, map[string]int) {
	return m.ips, m.count
}
func (m *mockRateLimiterForMetrics) GetRateLimitConfig() (*entity.RateLimitConfig, error) {
	return nil, nil
}
func (m *mockRateLimiterForMetrics) GetSubnetLimits(ip string) (*entity.SubnetRateLimit, error) {
	return nil, nil
}
func (m *mockRateLimiterForMetrics) CheckRequestLimit(ip string) (bool, string, error) {
	return true, "", nil
}
func (m *mockRateLimiterForMetrics) CheckTrafficLimit(ip string, d, u int64) (bool, string, error) {
	return true, "", nil
}
func (m *mockRateLimiterForMetrics) CheckConnectionLimit(ip string, inc bool) (bool, string, error) {
	return true, "", nil
}

func TestMetricsPuller_FullCoverage(t *testing.T) {
	mockList := &mockListUseCaseForMetrics{
		white: []string{"1.1.1.1", "2.2.2.2"},
		black: []string{"9.9.9.9"},
		gray:  []string{"10.0.0.1", "10.0.0.2"},
	}

	mockRL := &mockRateLimiterForMetrics{
		ips: []string{"1.1.1.1", "2.2.2.2", "3.3.3.3", "4.4.4.4", "5.5.5.5"},
		count: map[string]int{
			"1.1.1.1": 5000,
			"2.2.2.2": 3200,
			"3.3.3.3": 800,
			"4.4.4.4": 150,
			"5.5.5.5": 10,
		},
	}

	puller := CreateMetricsPuller(mockList, mockRL)

	t.Run("Init_LoadIpsLists", func(t *testing.T) {
		puller.Init()
		time.Sleep(400 * time.Millisecond)
		puller.Stop()
	})

	t.Run("UpdateTopUsers_MultipleItems", func(t *testing.T) {
		puller = CreateMetricsPuller(mockList, mockRL)
		puller.Init()
		time.Sleep(400 * time.Millisecond)
		puller.Stop()
	})

	t.Run("UpdateTopUsers_SingleItem", func(t *testing.T) {
		singleRL := &mockRateLimiterForMetrics{
			ips:   []string{"only.one.ip"},
			count: map[string]int{"only.one.ip": 9999},
		}
		p := CreateMetricsPuller(mockList, singleRL)
		p.Init()
		time.Sleep(300 * time.Millisecond)
		p.Stop()
	})

	t.Run("UpdateTopUsers_Empty", func(t *testing.T) {
		emptyRL := &mockRateLimiterForMetrics{ips: []string{}, count: map[string]int{}}
		p := CreateMetricsPuller(mockList, emptyRL)
		p.Init()
		time.Sleep(300 * time.Millisecond)
		p.Stop()
	})
}

func TestMetricsPuller_LoadIpsLists_Direct(t *testing.T) {
	mockList := &mockListUseCaseForMetrics{
		white: []string{"1.1.1.1", "2.2.2.2", "3.3.3.3"},
		black: []string{},
		gray:  []string{"10.0.0.1"},
	}

	mockRL := &mockRateLimiterForMetrics{ips: []string{}, count: map[string]int{}}

	puller := CreateMetricsPuller(mockList, mockRL)

	impl := puller.(*MetricsPullerImpl)
	impl.LoadIpsLists()

	t.Log("LoadIpsLists executed directly")
}
