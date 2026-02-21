package main

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pion/turn/v4"
)

func generateTURNCredentials(secret string, ttl time.Duration) (username, password string) {
	expiry := time.Now().Add(ttl).Unix()
	username = fmt.Sprintf("%d", expiry)
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(username))
	password = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return
}

func startTURN(cfg *Config) (*turn.Server, error) {
	if cfg.TURNPort == "" {
		return nil, fmt.Errorf("TURN port is required")
	}

	publicIp := cfg.PublicIp
	if publicIp == "" {
		publicIp = "127.0.0.1"
	}
	relayIp := net.ParseIP(publicIp)
	if relayIp == nil {
		resolved, err := net.ResolveIPAddr("ip4", publicIp)
		if err != nil || resolved == nil || resolved.IP == nil {
			return nil, fmt.Errorf("failed to resolve TURN relay IP from %q: %v", publicIp, err)
		}
		relayIp = resolved.IP
	}
	addr := net.JoinHostPort("0.0.0.0", cfg.TURNPort)
	udpListener, err := net.ListenPacket("udp4", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen UDP for TURN: %w", err)
	}

	tcpListener, err := net.Listen("tcp4", addr)
	if err != nil {
		_ = udpListener.Close()
		return nil, fmt.Errorf("failed to listen TCP for TURN: %w", err)
	}

	relayGen := &turn.RelayAddressGeneratorPortRange{
		RelayAddress: relayIp,
		Address:      "0.0.0.0",
		MinPort:      cfg.RelayPortMin,
		MaxPort:      cfg.RelayPortMax,
	}

	listenerConfigs := []turn.ListenerConfig{
		{
			Listener:              tcpListener,
			RelayAddressGenerator: relayGen,
		},
	}

	turnserver, err := turn.NewServer(turn.ServerConfig{
		Realm: cfg.TURNRealm,
		AuthHandler: func(username, realm string, srcAddr net.Addr) ([]byte, bool) {
			t, parseErr := strconv.ParseInt(username, 10, 64)
			if parseErr != nil {
				log.Printf("TURN auth: invalid username %q", username)
				return nil, false
			}
			if time.Now().Unix() > t {
				log.Printf("TURN auth: expired credential for %q", username)
				return nil, false
			}
			mac := hmac.New(sha1.New, []byte(cfg.TURNSecret))
			mac.Write([]byte(username))
			password := base64.StdEncoding.EncodeToString(mac.Sum(nil))

			h := md5.New()
			h.Write([]byte(username + ":" + realm + ":" + password))
			return h.Sum(nil), true
		},
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn:            udpListener,
				RelayAddressGenerator: relayGen,
			},
		},
		ListenerConfigs: listenerConfigs,
	})
	if err != nil {
		_ = udpListener.Close()
		_ = tcpListener.Close()
		return nil, fmt.Errorf("failed to create TURN server: %w", err)
	}

	log.Printf("TURN server listening on UDP+TCP :%s (realm=%s, publicIP=%s, relay=%d-%d)",
		cfg.TURNPort, cfg.TURNRealm, publicIp, cfg.RelayPortMin, cfg.RelayPortMax)
	return turnserver, nil

}

func getIceHost(cfg *Config, r *http.Request) string {
	if cfg.PublicHost != "" {
		return cfg.PublicHost
	}
	if cfg.PublicIp != "" {
		return cfg.PublicIp
	}
	if r != nil {
		if h := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Host"), ",")[0]); h != "" {
			if host, _, err := net.SplitHostPort(h); err == nil && host != "" {
				return host
			}
			return h
		}
		if r.Host != "" {
			if host, _, err := net.SplitHostPort(r.Host); err == nil && host != "" {
				return host
			}
			return r.Host
		}
	}
	return "127.0.0.1"
}

func buildIceServers(cfg *Config, r *http.Request, includeTURN bool) []IceServerInfo {
	if !includeTURN {
		return []IceServerInfo{{URLs: []string{"stun:stun.l.google.com:19302"}}}
	}

	host := getIceHost(cfg, r)
	if host == "localhost" {
		host = "127.0.0.1"
	}

	turnHostPort := net.JoinHostPort(host, cfg.TURNPort)
	stunURLs := []string{fmt.Sprintf("stun:%s", turnHostPort)}
	turnURLs := []string{
		fmt.Sprintf("turn:%s?transport=udp", turnHostPort),
		fmt.Sprintf("turn:%s?transport=tcp", turnHostPort),
	}

	servers := []IceServerInfo{{URLs: stunURLs}}

	if cfg.TURNSecret == "" || cfg.TURNRealm == "" {
		return servers
	}

	turnUser, turnPass := generateTURNCredentials(cfg.TURNSecret, 24*time.Hour)
	servers = append(servers, IceServerInfo{
		URLs:       turnURLs,
		Username:   turnUser,
		Credential: turnPass,
	})

	return servers
}
