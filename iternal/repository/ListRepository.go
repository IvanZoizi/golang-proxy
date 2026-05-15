package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"proxy/iternal/entity"
)

type ListRepository interface {
	GetAllIps() ([]string, error)
	GetAllCIDRIps() ([]string, error)
	GetAllRangeIps() ([]string, error)
	AddIp(ip string) error
	DeleteIp(ip string) error
	Contains(ip string) bool
}

type ListRepositoryImpl struct {
	list     *entity.List
	filename string
}

func NewListRepository(filename string) (ListRepository, error) {
	list := &entity.List{Ips: []string{}}

	if _, err := os.Stat(filename); err == nil {
		file, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		var data struct {
			Ips []string `json:"ips"`
		}

		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&data); err != nil {
			return nil, err
		}

		list.Ips = data.Ips
	}

	return &ListRepositoryImpl{
		list:     list,
		filename: filename,
	}, nil
}

func (r *ListRepositoryImpl) save() error {
	file, err := os.Create(r.filename)
	if err != nil {
		return err
	}
	defer file.Close()

	data := struct {
		Ips []string `json:"ips"`
	}{
		Ips: r.list.Ips,
	}

	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}

func (r *ListRepositoryImpl) GetAllIps() ([]string, error) {
	return r.list.GetAll(), nil
}

func (r *ListRepositoryImpl) AddIp(ip string) error {
	if !r.list.AddIp(ip) {
		return fmt.Errorf("ip %s already exists", ip)
	}
	return r.save()
}

func (r *ListRepositoryImpl) DeleteIp(ip string) error {
	if !r.list.RemoveIp(ip) {
		return fmt.Errorf("ip %s not found", ip)
	}
	return r.save()
}

func (r *ListRepositoryImpl) Contains(ip string) bool {
	return r.list.Contains(ip)
}

func (r *ListRepositoryImpl) GetAllCIDRIps() ([]string, error) {
	return r.list.GetAllCIDR(), nil
}

func (r *ListRepositoryImpl) GetAllRangeIps() ([]string, error) {
	return r.list.GetAllRangeIps(), nil
}
