package sync

import (
	"math/rand"
	"sync"
	"time"
)

// DefaultPeerPool implementa a interface PeerPool controlando
// nós conectados para o mecanismo de Fast Sync.
type DefaultPeerPool struct {
	mu    sync.RWMutex
	peers map[string]Peer
}

// NewPeerPool inicializa um PeerPool vazio.
func NewPeerPool() *DefaultPeerPool {
	return &DefaultPeerPool{
		peers: make(map[string]Peer),
	}
}

// AddPeer insere ou sobrecreve um peer no pool.
func (p *DefaultPeerPool) AddPeer(peer Peer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers[peer.ID()] = peer
}

// RemovePeer desconecta (caso aplicável) e remove o peer do pool.
func (p *DefaultPeerPool) RemovePeer(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer, ok := p.peers[id]; ok {
		peer.Disconnect() // Encapsulamento
		delete(p.peers, id)
	}
}

// PeerCount retorna a quantidade de peers ativos no pool.
func (p *DefaultPeerPool) PeerCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.peers)
}

// ListPeers retorna a cópia da lista de peers para iteração segura.
func (p *DefaultPeerPool) ListPeers() []Peer {
	p.mu.RLock()
	defer p.mu.RUnlock()

	list := make([]Peer, 0, len(p.peers))
	for _, peer := range p.peers {
		list = append(list, peer)
	}
	return list
}

// BestPeer executa o Algoritmo de Seleção para encontrar o Peer mais apto.
// Score = (height * 4) - latency (ms) - (failures * 100) + connectionTimeBonus (s)
func (p *DefaultPeerPool) BestPeer() Peer {
	p.mu.Lock()
	defer p.mu.Unlock()

	var best Peer
	var maxScore float64 = -999999999.0 // score inicial baixo

	toRemove := []string{}

	for id, peer := range p.peers {
		if peer.Failures() > 3 { // Threshold para considerar indisponível
			toRemove = append(toRemove, id)
			continue
		}

		score := calculateScore(peer)
		if best == nil || score > maxScore {
			maxScore = score
			best = peer
		}
	}

	for _, id := range toRemove {
		if peer, ok := p.peers[id]; ok {
			peer.Disconnect()
			delete(p.peers, id)
		}
	}

	return best
}

// RandomPeer seleciona um Peer aleatório.
func (p *DefaultPeerPool) RandomPeer() Peer {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.peers) == 0 {
		return nil
	}

	idx := rand.Intn(len(p.peers))
	i := 0
	for _, peer := range p.peers {
		if i == idx {
			return peer
		}
		i++
	}
	return nil
}

// FastestPeer ignora outras variáveis e pega a menor latência.
func (p *DefaultPeerPool) FastestPeer() Peer {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var fastest Peer
	for _, peer := range p.peers {
		if fastest == nil || peer.Latency() < fastest.Latency() {
			fastest = peer
		}
	}
	return fastest
}

// HighestPeer ignora outras variáveis e pega o maior Height.
func (p *DefaultPeerPool) HighestPeer() Peer {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var highest Peer
	for _, peer := range p.peers {
		if highest == nil || peer.Height() > highest.Height() {
			highest = peer
		}
	}
	return highest
}

// calculateScore encapsula a matemática do algoritmo de seleção de Peers.
func calculateScore(p Peer) float64 {
	heightBonus := float64(p.Height() * 4)
	latencyPenal := float64(p.Latency().Milliseconds())
	failuresPenal := float64(p.Failures() * 100)
	uptimeBonus := time.Since(p.ConnectedSince()).Seconds()

	return heightBonus - latencyPenal - failuresPenal + uptimeBonus
}
