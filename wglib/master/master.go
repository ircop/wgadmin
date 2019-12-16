package master

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/centrifugal/centrifuge"
	"github.com/gogo/protobuf/proto"
	wgproto "github.com/ircop/wgadmin/wglib/proto"
	"go.uber.org/zap"
)

const keyCtxRemoteAddr = iota

type Master struct {
	logger *zap.Logger

	clients sync.Map
	tasks   sync.Map

	taskTimeout time.Duration
}

// NewMasterDaemon creates and returns new Master instance
// logger: instance of zap.Logger. taskTimeout: timeout for waiting of slave tasks execution.
func NewMasterDaemon(logger *zap.Logger, taskTimout time.Duration) *Master {
	return &Master{
		logger:      logger,
		taskTimeout: taskTimout,
	}
}

// Run initializes centrifuge instance; creates and returns websocket http handler
func (m *Master) Run() (http.Handler, error) {
	cfg := centrifuge.DefaultConfig
	cfg.Publish = true
	cfg.LogLevel = centrifuge.LogLevelInfo
	cfg.LogHandler = func(e centrifuge.LogEntry) {
		m.logger.Info("centrifuge", zap.String("message", e.Message),
			zap.String("fields", fmt.Sprintf("%+v", e.Fields)))
	}
	cfg.ClientInsecure = true
	cfg.Secret = "wgadmin"

	node, err := centrifuge.New(cfg)
	if err != nil {
		return nil, err
	}

	node.On().ClientConnected(m.clientConnected)

	m.logger.Info("starting centrifuge listener")

	if err = node.Run(); err != nil {
		return nil, err
	}

	return handleWS(centrifuge.NewWebsocketHandler(node, centrifuge.WebsocketConfig{})), nil
}

func (m *Master) clientConnected(ctx context.Context, client *centrifuge.Client) {
	m.logger.Info("client connected", zap.String("remote", ctx.Value(keyCtxRemoteAddr).(string)),
		zap.String("transport", client.Transport().Name()), zap.String("id", client.ID()),
		zap.String("proto", fmt.Sprintf("%v", client.Transport().Protocol())))

	m.clients.Store(ctx.Value(keyCtxRemoteAddr), client)

	client.On().Disconnect(func(ev centrifuge.DisconnectEvent) centrifuge.DisconnectReply {
		m.logger.Info("client disconnected", zap.String("cliend-id", client.ID()),
			zap.String("event", fmt.Sprintf("%+#v", ev)))

		m.clients.Delete(client.ID())

		return centrifuge.DisconnectReply{}
	})

	client.On().Message(func(ev centrifuge.MessageEvent) centrifuge.MessageReply {
		m.logger.Info("got client message", zap.String("client", ctx.Value(keyCtxRemoteAddr).(string)))

		m.processIncomingMessage(ctx, ev.Data)

		return centrifuge.MessageReply{}
	})
}

func (m *Master) processIncomingMessage(ctx context.Context, message centrifuge.Raw) {
	var msg wgproto.WGPacket

	peer := ctx.Value(keyCtxRemoteAddr).(string)

	err := proto.Unmarshal(message, &msg)
	if err != nil {
		m.logger.Info("failed to unmarshal packet", zap.String("peer", peer),
			zap.String("remote", ctx.Value(keyCtxRemoteAddr).(string)), zap.Error(err))
		return
	}

	m.logger.Info("got client message", zap.String("peer", peer), zap.String("id", msg.UUID),
		zap.String("type", msg.PacketType.String()))

	if msg.UUID == "" {
		m.logger.Info("ignoring message with empty id")
		return
	}

	taskChan, ok := m.tasks.Load(msg.UUID)
	if !ok {
		m.logger.Info("task not found", zap.String("id", msg.UUID))
		return
	}

	taskChan.(chan wgproto.WGPacket) <- msg
}

//func (m *Master) getPeer(peer string) (*centrifuge.Client) {
func (m *Master) getPeer(peer string) Sender {
	if net.ParseIP(peer) == nil {
		return nil
	}

	clientIf, ok := m.clients.Load(peer)
	if !ok {
		return nil
	}
	//client := clientIf.(*centrifuge.Client)
	client := clientIf.(Sender)

	return client
}

func handleWS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		ctx := context.WithValue(context.Background(), keyCtxRemoteAddr, host) //nolint:golint

		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	})
}
