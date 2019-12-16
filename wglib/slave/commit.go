package slave

import (
	"os"
	"strings"
)

// commit saves current peer-config. Server config is based on predefined template.
func (s *Slave) Commit() error {
	if !s.autosave {
		return nil
	}

	var content strings.Builder

	content.WriteString(s.config.SaveTemplate)

	peers, err := s.peerConfig()
	if err != nil {
		return err
	}

	content.WriteString(peers)

	// write file
	f, err := os.OpenFile(s.config.SavePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	_, err = f.WriteString(content.String())
	if err != nil {
		return err
	}

	return nil
}

func (s *Slave) peerConfig() (string, error) {
	var content strings.Builder

	dev, err := s.wg.Device(s.ifname)
	if err != nil {
		return "", err
	}

	for _, peer := range dev.Peers {
		content.WriteString("[Peer]\n")
		content.WriteString("PublicKey = ")
		content.WriteString(peer.PublicKey.String())
		content.WriteString("\nAllowedIps = ")
		content.WriteString(peer.AllowedIPs[0].String())
		content.WriteString("\n\n")
	}

	return content.String(), nil
}
