package slave

import (
	"crypto/tls"
	"encoding/base64"
	"sync"
	"time"

	"github.com/centrifugal/centrifuge-go"
	"github.com/gogo/protobuf/proto"
	wgproto "github.com/pnforge/wgadmin/wglib/proto"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
)

type Config struct {
	RemoteAddr     string
	BasicAuthLogin string
	BasicAuthPW    string
	SaveTemplate   string
	SavePath       string
	SkipTLSVerify  bool
}

type Slave struct {
	logger  *zap.Logger
	ifname  string
	wg      *wgctrl.Client
	packets chan wgproto.WGPacket

	config   Config
	autosave bool

	sync.Mutex
}

func NewSLave(ifname string, logger *zap.Logger) (*Slave, error) {
	var err error

	daemon := &Slave{
		logger: logger,
		ifname: ifname,
	}

	// init wireguard
	daemon.wg, err = wgctrl.New()
	if err != nil {
		return nil, err
	}

	return daemon, nil
}

// Run wgslave daemon
func (s *Slave) Run(config Config) error {
	s.logger.Info("running wgslave daemon", zap.String("remote", config.RemoteAddr))

	cfg := centrifuge.DefaultConfig()
	cfg.EnableCompression = true
	cfg.PingInterval = time.Second * 3
	cfg.WriteTimeout = time.Second * 3

	if config.BasicAuthLogin != "" && config.BasicAuthPW != "" {
		auth := config.BasicAuthLogin + ":" + config.BasicAuthPW
		auth = base64.StdEncoding.EncodeToString([]byte(auth))
		cfg.Header.Add("Authorization", "Basic "+auth)
	}

	if config.SkipTLSVerify {
		// nolint:gosec
		cfg.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	s.config = config
	if s.config.SavePath != "" && s.config.SaveTemplate != "" {
		s.autosave = true
	}

	s.packets = make(chan wgproto.WGPacket)

	client := centrifuge.New("wss://"+config.RemoteAddr+"?format=protobuf", cfg)
	client.OnConnect(s)
	client.OnDisconnect(s)
	client.OnError(s)
	client.OnMessage(s)

	if err := client.Connect(); err != nil {
		s.logger.Info("failed to connect to wgmaster", zap.Error(err))
	}

	for packet := range s.packets {
		data, err := proto.Marshal(&packet) // nolint:scopelint
		if err != nil {
			s.logger.Error("failed to marshal packet", zap.String("type", packet.PacketType.String()), zap.Error(err))
			continue
		}

		if err = client.Send(data); err != nil {
			s.logger.Error("failed to send message to wgmaster", zap.Error(err))
		}
	}

	select {}
}

func (s *Slave) OnConnect(client *centrifuge.Client, ev centrifuge.ConnectEvent) {
	// nolint:godox
	// todo: request sync?
	s.logger.Info("connected to wgmaster")
}
func (s *Slave) OnDisconnect(client *centrifuge.Client, ev centrifuge.DisconnectEvent) {
	s.logger.Info("disconnected from wgmaster")
}
func (s *Slave) OnError(client *centrifuge.Client, ev centrifuge.ErrorEvent) {
	s.logger.Info("centerifuge error", zap.String("message", ev.Message))
}
func (s *Slave) OnMessage(client *centrifuge.Client, ev centrifuge.MessageEvent) {
	s.logger.Info("message from wgmaster")

	var msg wgproto.WGPacket
	if err := proto.Unmarshal(ev.Data, &msg); err != nil {
		s.logger.Error("failed to unmarshal message", zap.Error(err))
		return
	}

	switch msg.PacketType {
	case wgproto.PacketType_PT_IFS_REQUEST:
		s.interfaces(msg.UUID)
	case wgproto.PacketType_PT_ADD_IF:
		s.addIfs(msg)
	case wgproto.PacketType_PT_REMOVE_IF:
		s.removeIfs(msg)
	case wgproto.PacketType_PT_SYNC_RESPONSE:
		s.sync(msg)
	/*
		case wgproto.PacketType_PT_IFS_REQUEST:
			s.logger.Info("parsed into interfaces request")
			s.interfaces(msg.UUID)
	*/
	default:
		s.logger.Info("unknown message type", zap.String("type", msg.PacketType.String()))
		s.sendError(msg.UUID, "wgslave: not implemented")
	}
}
