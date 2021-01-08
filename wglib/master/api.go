package master

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	wgproto "github.com/pnforge/wgadmin/wglib/proto"
)

type Interface struct {
	Key string `json:"key"`
	IP  string `json:"ip"`
}

var ErrPeerNodFound = errors.New("peer not found")
var ErrTaskTimeout = errors.New("task timed out")
var ErrWrongParams = errors.New("wrong input parameters")

// Peers returns list of ip-addresses of connected wgslave peers
func (m *Master) Peers() []string {
	peers := make([]string, 0)

	m.clients.Range(func(k, v interface{}) bool {
		peers = append(peers, k.(string))

		return true
	})

	return peers
}

// Peer interfaces returns lits of wg interfaces of specified peer, or error
func (m *Master) PeerInterfaces(peer string) ([]Interface, error) {
	client := m.getPeer(peer)
	if client == nil {
		return nil, ErrPeerNodFound
	}

	requestID := uuid.New().String()
	resultChan := make(chan wgproto.WGPacket, 1)

	m.tasks.Store(requestID, resultChan)

	if err := m.sendPacket(requestID, wgproto.PacketType_PT_IFS_REQUEST, client); err != nil {
		m.tasks.Delete(requestID)
		return nil, fmt.Errorf("PeerInterfaces [%s]: error sending packet: %w", peer, err)
	}

	ifs := make([]Interface, 0)

	result, err := m.waitResult(requestID)
	if err != nil {
		return nil, fmt.Errorf("PeerInterfaces [%s] failed: %w", peer, err)
	}

	if result.PacketType == wgproto.PacketType_PT_ERROR {
		return nil, fmt.Errorf("PeerInterfaces [%s] failed: %s", peer, result.Error)
	}

	var resultPacket wgproto.Interfaces
	if err = proto.Unmarshal(result.Payload.Value, &resultPacket); err != nil {
		return nil, fmt.Errorf("PeerInterfaces [%s]: result unmarshal failed: %w", peer, err)
	}

	for i := range resultPacket.Interfaces {
		ifs = append(ifs, Interface{
			Key: resultPacket.Interfaces[i].PubKey,
			IP:  resultPacket.Interfaces[i].IP,
		})
	}

	return ifs, nil
}

// DelInterfacesAllPeers sends del-interface request to all connected peers
func (m *Master) DelInterfacesAllPeers(keys []string) []error {
	errs := []error{}

	var mu sync.Mutex

	var wg sync.WaitGroup

	m.clients.Range(func(k, v interface{}) bool {
		peer := k.(string)

		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			err := m.DelInterfacesPeer(peer, keys)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}

			wg.Done()
		}(&wg)

		return true
	})

	wg.Wait()

	return errs
}

// PeerDelInterface sends del-interface request to specified peer
func (m *Master) DelInterfacesPeer(peer string, keys []string) error {
	if len(keys) == 0 {
		return ErrWrongParams
	}

	client := m.getPeer(peer)
	if client == nil {
		return ErrPeerNodFound
	}

	requestID := uuid.New().String()
	resultChan := make(chan wgproto.WGPacket, 1)

	m.tasks.Store(requestID, resultChan)

	if err := m.sendDelRequest(requestID, keys, client); err != nil {
		m.tasks.Delete(requestID)
		return fmt.Errorf("DelInterfacesPeer [%s]: error sending packet: %w", peer, err)
	}

	result, err := m.waitResult(requestID)
	if err != nil {
		return fmt.Errorf("DelInterfacesPeer [%s] failed: %w", peer, err)
	}

	if result.PacketType == wgproto.PacketType_PT_ERROR {
		return fmt.Errorf("DelInterfacesPeer [%s] failed: %s", peer, result.Error)
	}

	var resultPacket wgproto.Result
	if err = proto.Unmarshal(result.Payload.Value, &resultPacket); err != nil {
		return fmt.Errorf("DelInterfacesPeer [%s]: result unmarshal failed: %w", peer, err)
	}

	if !resultPacket.Success {
		return fmt.Errorf("DelInterfacesPeer [%s]: wgslave failed: %s", peer, resultPacket.Error)
	}

	return nil
}

// PutInterfacesAllPeers sends put-request to all connected peers
func (m *Master) PutInterfacesAllPeers(interfaces []Interface, syncPacket bool) []error {
	errs := []error{}

	var mu sync.Mutex

	var wg sync.WaitGroup

	m.clients.Range(func(k, v interface{}) bool {
		peer := k.(string)

		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			err := m.PutInterfacesPeer(peer, interfaces, syncPacket)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
			wg.Done()
		}(&wg)

		return true
	})

	wg.Wait()

	return errs
}

// PeerPut sends request to add new interface to peer
func (m *Master) PutInterfacesPeer(peer string, interfaces []Interface, syncPacket bool) error {
	if len(interfaces) == 0 {
		return ErrWrongParams
	}

	client := m.getPeer(peer)
	if client == nil {
		return ErrPeerNodFound
	}

	requestID := uuid.New().String()
	resultChan := make(chan wgproto.WGPacket, 1)

	m.tasks.Store(requestID, resultChan)

	if err := m.sendPutRequest(requestID, interfaces, client, syncPacket); err != nil {
		m.tasks.Delete(requestID)
		return fmt.Errorf("PeerPutInterface [%s]: error sending resultPacket: %w", peer, err)
	}

	result, err := m.waitResult(requestID)
	if err != nil {
		return fmt.Errorf("PutInterfacesPeer [%s] failed: %w", peer, err)
	}

	if result.PacketType == wgproto.PacketType_PT_ERROR {
		return fmt.Errorf("PutInterfacesPeer [%s] failed: %s", peer, result.Error)
	}

	var resultPacket wgproto.Result
	if err = proto.Unmarshal(result.Payload.Value, &resultPacket); err != nil {
		return fmt.Errorf("PutInterfacesPeer [%s]: result unmarshal failed: %w", peer, err)
	}

	if !resultPacket.Success {
		return fmt.Errorf("PutInterfacesPeer [%s]: wgslave failed: %s", peer, resultPacket.Error)
	}

	return nil
}

func (m *Master) waitResult(requestID string) (wgproto.WGPacket, error) {
	resultChan, ok := m.tasks.Load(requestID)
	if !ok {
		return wgproto.WGPacket{}, fmt.Errorf("result channel not found")
	}

	t := time.NewTimer(m.taskTimeout)

	select {
	case <-t.C:
		m.tasks.Delete(requestID)
		return wgproto.WGPacket{}, ErrTaskTimeout
	case result := <-resultChan.(chan wgproto.WGPacket):
		m.tasks.Delete(requestID)
		return result, nil
	}
}
