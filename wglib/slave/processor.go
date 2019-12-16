package slave

import (
	"net"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	wgproto "github.com/ircop/wgadmin/wglib/proto"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func (s *Slave) interfaces(id string) {
	dev, err := s.wg.Device(s.ifname)
	if err != nil {
		s.sendError(id, "wg device error: "+err.Error())
		return
	}

	var interfaces wgproto.Interfaces
	interfaces.Interfaces = make([]*wgproto.Interface, 0)

	s.Lock()
	for i := range dev.Peers {
		var iface wgproto.Interface
		iface.PubKey = dev.Peers[i].PublicKey.String()

		if len(dev.Peers[i].AllowedIPs) > 0 {
			iface.IP = dev.Peers[i].AllowedIPs[0].IP.String()
		}

		interfaces.Interfaces = append(interfaces.Interfaces, &iface)
	}
	s.Unlock()

	data, err := proto.Marshal(&interfaces)
	if err != nil {
		s.sendError(id, "packet marshaling error: "+err.Error())
		return
	}

	var msg wgproto.WGPacket
	msg.UUID = id
	msg.PacketType = wgproto.PacketType_PT_IFS_RESPONSE
	msg.Payload = &any.Any{TypeUrl: wgproto.PacketType_PT_IFS_RESPONSE.String(), Value: data}

	s.packets <- msg
}

func (s *Slave) addIfs(msg wgproto.WGPacket) {
	s.logger.Info("addIfs request")

	var ifs wgproto.Interfaces

	err := proto.Unmarshal(msg.Payload.Value, &ifs)
	if err != nil {
		s.sendError(msg.UUID, "interfaces unmarshaling error: "+err.Error())
		return
	}

	peers := []wgtypes.PeerConfig{}

	for _, iface := range ifs.Interfaces {
		pubkey, err := wgtypes.ParseKey(iface.PubKey)
		if err != nil {
			s.sendError(msg.UUID, "wrong interface key: "+iface.PubKey)
			return
		}

		_, ipnet, err := net.ParseCIDR(iface.IP + "/32")
		if err != nil {
			s.sendError(msg.UUID, "wrong interface ip: "+iface.IP)
			return
		}

		peers = append(peers, wgtypes.PeerConfig{
			PublicKey:         pubkey,
			ReplaceAllowedIPs: true,
			AllowedIPs:        []net.IPNet{*ipnet},
		})
	}

	newConfig := wgtypes.Config{
		ReplacePeers: false,
		Peers:        peers,
	}

	s.Lock()
	defer s.Unlock()

	if err = s.wg.ConfigureDevice(s.ifname, newConfig); err != nil {
		s.logger.Info("addIfs: failed to ConfigureDevice", zap.String("interface", s.ifname), zap.Error(err))
		s.sendError(msg.UUID, err.Error())

		return
	}

	s.sendOK(msg.UUID)

	if err = s.Commit(); err != nil {
		s.logger.Error("failed to commit config", zap.Error(err))
	}
}

func (s *Slave) removeIfs(msg wgproto.WGPacket) {
	s.logger.Info("removeIfs request")

	var req wgproto.RemoveRequest

	err := proto.Unmarshal(msg.Payload.Value, &req)
	if err != nil {
		s.sendError(msg.UUID, "keys unmarshaling error: "+err.Error())
		return
	}

	peers := []wgtypes.PeerConfig{}

	for _, key := range req.Keys {
		pubkey, err := wgtypes.ParseKey(key)
		if err != nil {
			s.sendError(msg.UUID, "wrong interface key: "+key)
			return
		}

		peers = append(peers, wgtypes.PeerConfig{
			PublicKey: pubkey,
			Remove:    true,
		})
	}

	newConfig := wgtypes.Config{
		ReplacePeers: false,
		Peers:        peers,
	}

	s.Lock()
	defer s.Unlock()

	if err = s.wg.ConfigureDevice(s.ifname, newConfig); err != nil {
		s.logger.Info("removeIfs: failed to ConfigureDevice", zap.String("interface", s.ifname), zap.Error(err))
		s.sendError(msg.UUID, err.Error())

		return
	}

	s.sendOK(msg.UUID)

	if err = s.Commit(); err != nil {
		s.logger.Error("failed to commit config", zap.Error(err))
	}
}

// Synchronize actual wireguard peers with interfaces in incoming message.
// Delete non-existing peers and add new ones.
func (s *Slave) sync(msg wgproto.WGPacket) {
	s.logger.Info("sync request")

	var newIfs wgproto.Interfaces

	if err := proto.Unmarshal(msg.Payload.Value, &newIfs); err != nil {
		s.logger.Error("sync: failed to unmarshal interfaces", zap.Error(err))
		s.sendError(msg.UUID, "sync: interfaces unmarshal error: "+err.Error())

		return
	}

	s.Lock()
	defer s.Unlock()

	/*
		just create new config, with fully replacing old one
	*/
	peers := []wgtypes.PeerConfig{}

	for _, iface := range newIfs.Interfaces {
		pubkey, err := wgtypes.ParseKey(iface.PubKey)
		if err != nil {
			s.sendError(msg.UUID, "wrong interface key: "+iface.PubKey)
			return
		}

		_, ipnet, err := net.ParseCIDR(iface.IP + "/32")
		if err != nil {
			s.sendError(msg.UUID, "wrong interface ip: "+iface.IP)
			return
		}

		peers = append(peers, wgtypes.PeerConfig{
			PublicKey:  pubkey,
			UpdateOnly: false,
			AllowedIPs: []net.IPNet{*ipnet},
		})
	}

	newConfig := wgtypes.Config{
		ReplacePeers: true,
		Peers:        peers,
	}

	if err := s.wg.ConfigureDevice(s.ifname, newConfig); err != nil {
		s.logger.Info("sync: failed to ConfigureDevice", zap.Error(err))
		s.sendError(msg.UUID, err.Error())

		return
	}

	s.sendOK(msg.UUID)

	if err := s.Commit(); err != nil {
		s.logger.Error("failed to commit config", zap.Error(err))
	}
}

func (s *Slave) sendOK(id string) {
	s.logger.Info("sending OK packet", zap.String("packet-id", id))

	var msg wgproto.WGPacket

	var result wgproto.Result

	msg.UUID = id
	msg.PacketType = wgproto.PacketType_PT_RESULT
	result.Success = true

	data, err := proto.Marshal(&result)
	if err != nil {
		s.logger.Error("failed to marshal OK result", zap.String("packet-id", id), zap.Error(err))
		return
	}

	msg.Payload = &any.Any{
		TypeUrl: wgproto.PacketType_PT_RESULT.String(),
		Value:   data,
	}

	s.packets <- msg
}

func (s *Slave) sendError(id string, errstr string) {
	s.logger.Info("sending error packet", zap.String("packet-id", id), zap.String("error", errstr))

	var msg wgproto.WGPacket

	msg.PacketType = wgproto.PacketType_PT_ERROR
	msg.Error = errstr
	msg.UUID = id

	s.packets <- msg
}
