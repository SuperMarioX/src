package service

import (
	"github.com/GoCollaborate/artifacts/card"
	"github.com/GoCollaborate/artifacts/parameter"
	"github.com/GoCollaborate/constants"
	"math/rand"
	"sync"
	"time"
)

type Service struct {
	ServiceID   string                `json:"serviceid,omitempty"`
	Description string                `json:"description"`
	Parameters  []parameter.Parameter `json:"parameters"`
	RegList     []card.Card           `json:"registers"`
	// a map of last heartbeat timestamps, eact key corresponds to one particular card full endpoint
	Heartbeats       map[string]int64 `json:"heartbeats,omitempty"`
	SbscrbList       []string         `json:"subscribers"`  // subscriber tokens
	Dependencies     []string         `json:"dependencies"` // dependent ServiceIDs
	Mode             Mode             `json:"mode,omitempty"`
	LoadBalanceMode  Mode             `json:"load_balance_mode,omitemtpy"`
	Version          string           `json:"version"`
	PlatformVersion  string           `json:"platform_version"`
	LastAssignedTo   card.Card        `json:"last_assigned_to,omitempty"`
	LastAssignedTime int64            `json:"last_assigned_time,omitempty"`
	serviceLock      sync.Mutex       `json:"-"`
}

func NewService() *Service {
	return &Service{
		Description:     "",
		Parameters:      []parameter.Parameter{},
		RegList:         []card.Card{},
		Heartbeats:      map[string]int64{},
		SbscrbList:      []string{},
		Dependencies:    []string{},
		Version:         "1.0",
		PlatformVersion: "golang1.8.1",
		serviceLock:     sync.Mutex{},
	}
}

func NewServiceFrom(from *Service) *Service {
	return &Service{
		Description:     from.Description,
		Parameters:      from.Parameters,
		RegList:         from.RegList,
		Heartbeats:      map[string]int64{},
		SbscrbList:      from.SbscrbList,
		Dependencies:    from.Dependencies,
		Version:         from.Version,
		PlatformVersion: from.PlatformVersion,
		serviceLock:     sync.Mutex{},
	}
}

func (s *Service) SetMode(m *Mode) Mode {
	s.Mode = *m
	return *m
}

func (s *Service) GetMode() Mode {
	return s.Mode
}

func (s *Service) GetDependencies() []string {
	return s.Dependencies
}

func (s *Service) SetLoadBalanceMode(m *Mode) Mode {
	s.LoadBalanceMode = *m
	return *m
}

func (s *Service) GetLoadBalanceMode() Mode {
	return s.LoadBalanceMode
}

func (s *Service) Register(agt *card.Card) error {
	s.serviceLock.Lock()
	defer s.serviceLock.Unlock()
	for _, r := range s.RegList {
		if agt.IsEqualTo(&r) {
			return constants.ErrConflictRegister
		}
	}
	s.RegList = append(s.RegList, *agt)
	return nil
}

// de-register function will not preserve the original order of registers
func (s *Service) DeRegister(agt *card.Card) error {
	s.serviceLock.Lock()
	defer s.serviceLock.Unlock()
	y := s.RegList[:0]

	for i, x := range s.RegList {
		if agt.IsEqualTo(&x) {
			y = append(s.RegList[:i], s.RegList[i+1:]...)
			s.RegList = y
			return nil
		}
	}
	return constants.ErrNoRegister
}

func (s *Service) DeRegisterAll() error {
	s.serviceLock.Lock()
	defer s.serviceLock.Unlock()
	y := []card.Card{}
	s.RegList = y
	return nil
}

func (s *Service) Subscribe(token string) error {
	s.serviceLock.Lock()
	defer s.serviceLock.Unlock()
	for _, sbscr := range s.SbscrbList {
		if token == sbscr {
			return constants.ErrConflictSubscriber
		}
	}
	s.SbscrbList = append(s.SbscrbList, token)
	return nil
}

func (s *Service) UnSubscribe(token string) error {
	s.serviceLock.Lock()
	defer s.serviceLock.Unlock()
	y := s.SbscrbList[:0]
	for i, x := range s.SbscrbList {
		if x == token {
			y = append(s.SbscrbList[:i], s.SbscrbList[i+1:]...)
			s.SbscrbList = y
			return nil
		}
	}
	return constants.ErrNoSubscriber
}

func (s *Service) Query(token string) (*card.Card, error) {
	for _, x := range s.SbscrbList {
		if x == token {
			return s.loadBalanceRoll(), nil
		}
	}
	return nil, constants.ErrNoRegister
}

func (s *Service) loadBalanceRoll() *card.Card {
	switch s.LoadBalanceMode {
	default:
		return &s.RegList[rand.Intn(len(s.RegList))]
	}
}

func (s *Service) UnSubscribeAll() error {
	s.serviceLock.Lock()
	defer s.serviceLock.Unlock()
	y := []string{}
	s.SbscrbList = y
	return nil
}

func (s *Service) Heartbeat(agt *card.Card) {
	s.serviceLock.Lock()
	defer s.serviceLock.Unlock()
	y := s.RegList
	for _, x := range y {
		if agt.IsEqualTo(&x) {
			s.Heartbeats[x.GetFullEndPoint()] = time.Now().Unix()
			return
		}
	}
}
