package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"pc_node/storage/jsonstorage"
	"pc_node/sync"
)

// RunFastSyncBootstrap orquestra a inicialização do Fast Sync da Nebula Network.
func RunFastSyncBootstrap(localHeight uint64) error {
	log.Println("🔄 Iniciando Bootstrap: Avaliando necessidade de Fast Sync...")

	pool := sync.NewPeerPool()

	// 1. Descobrir peers (mock para esta fase, já que a rede real não foi acoplada ao TCPPeer)
	discoverPeers(pool)

	if pool.PeerCount() == 0 {
		log.Println("⚠️ Nenhum peer encontrado. Iniciando nó normalmente.")
		return nil
	}

	remoteHeight := getHighestKnownHeight(pool)

	log.Printf("📊 Altura Local: %d | Altura Remota Máxima: %d\n", localHeight, remoteHeight)

	if remoteHeight <= localHeight {
		log.Println("✅ Nó atualizado. Iniciando normalmente.")
		return nil
	}

	log.Println("⚡ Nó desatualizado. Iniciando Fast Sync Pipeline...")

	manager := sync.NewSyncManager(pool)
	manager.UpdateLocalHeight(localHeight)

	queue := sync.NewDownloadQueue(3)

	// Utiliza concurrency = 5, timeout 10s
	downloader := sync.NewDownloader(queue, pool, 5, 10*time.Second)
	validator := sync.NewBlockValidator(sync.BlockValidatorEventHandlers{})

	// Integra com o storage real atual do Node
	engine := jsonstorage.NewJSONEngine()
	// Precisamos abrir o engine, e fechar no final
	if err := engine.Open("ledger.json"); err != nil {
		return fmt.Errorf("falha ao abrir engine de storage: %v", err)
	}
	defer engine.Close()

	importer := sync.NewBlockImporter(engine, sync.BlockImporterEventHandlers{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var syncErr error
	events := sync.SyncControllerEventHandlers{
		OnStateChanged: func(oldState, newState sync.SyncState) {
			log.Printf("Fast Sync Estado: %s -> %s\n", oldState, newState)
		},
		OnProgress: func(report sync.SyncStatusReport) {
			log.Printf("Progress: %d/%d (%.2f%%) - %d blocks/sec",
				report.ImportedBlocks, report.RemoteHeight, report.ProgressPercent, int(report.SpeedBlocksSec))
		},
		OnCompleted: func(report sync.SyncStatusReport) {
			log.Println("🎉 Fast Sync Concluído com sucesso!")
			cancel() // Termina a espera
		},
		OnFailed: func(err error) {
			syncErr = errors.New("falha no fast sync: " + err.Error())
			cancel()
		},
		OnPipelineError: func(err error) {
			syncErr = errors.New("erro no pipeline: " + err.Error())
			cancel()
		},
		OnCancelled: func() {
			syncErr = errors.New("sincronização cancelada pelo usuário")
			cancel()
		},
	}

	controller := sync.NewSyncController(manager, downloader, validator, importer, events)

	if err := controller.Start(remoteHeight); err != nil {
		return err
	}

	// Aguarda conclusão (OnCompleted ou falhas que dão cancel)
	<-ctx.Done()

	if syncErr != nil {
		// Parar tudo
		controller.Cancel()
		return syncErr
	}

	return nil
}

// discoverPeers descobre e conecta a peers.
func discoverPeers(pool sync.PeerPool) {
	// Em um nó real, nós leríamos os IPs de seed nodes ou de um cache local.
	// Como estamos rodando na mesma máquina localmente para testes, vamos tentar conectar no localhost.
	// Se tivermos seed nodes, iteraríamos por eles.
	seedNodes := []string{"127.0.0.1:5556"}

	for _, addr := range seedNodes {
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err == nil {
			peer := sync.NewTCPPeer(addr, conn, 0)
			pool.AddPeer(peer)
		}
	}
}

func getHighestKnownHeight(pool sync.PeerPool) uint64 {
	var max uint64 = 0
	for _, p := range pool.ListPeers() {
		// Obtém a altura real requisitando o status
		tcpPeer, ok := p.(*sync.TCPPeer)
		if ok {
			err := tcpPeer.GetStatus()
			if err == nil {
				// Aguarda a resposta de status
				msg, err := tcpPeer.Receive()
				if err == nil && msg.Type == sync.MsgTypeStatus {
					var statusMsg sync.SyncStatusMsg
					if err := msg.UnmarshalPayload(&statusMsg); err == nil {
						tcpPeer.UpdateHeight(statusMsg.Height)
					}
				}
			}
		}

		if p.Height() > max {
			max = p.Height()
		}
	}
	return max
}
