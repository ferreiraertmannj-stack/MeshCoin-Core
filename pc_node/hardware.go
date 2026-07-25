package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// DetectStorageType usa o PowerShell do Windows para verificar se o disco é SSD ou HDD
func DetectStorageType() string {
	// Comando para pegar o tipo da primeira mídia física do Windows
	cmdStr := `Get-PhysicalDisk | Select-Object MediaType | ConvertTo-Json`
	out, err := exec.Command("powershell", "-NoProfile", "-Command", cmdStr).Output()
	if err != nil {
		log.Println("Aviso: Falha ao detectar tipo de disco, assumindo HDD por padrão.", err)
		return "HDD"
	}

	// O output pode ser um array ou um único objeto. Vamos verificar por "SSD" na string crua por segurança
	outputStr := string(out)
	if strings.Contains(outputStr, `"SSD"`) {
		return "SSD"
	}
	
	return "HDD"
}

// AllocateNebulaCloudStorage aloca um arquivo 'sparse' no disco para garantir o espaço cedido à nuvem
func AllocateNebulaCloudStorage(gigabytes int) error {
	filename := "nebula_cloud_storage.dat"
	
	// Se já existe, apenas verifica o tamanho
	if stat, err := os.Stat(filename); err == nil {
		expectedBytes := int64(gigabytes) * 1024 * 1024 * 1024
		if stat.Size() >= expectedBytes {
			log.Printf("Nebula Cloud: Armazenamento de %d GB já alocado e detectado.\n", gigabytes)
			return nil
		}
		// Se for menor, deleta para realocar
		os.Remove(filename)
	}

	log.Printf("Nebula Cloud: Alocando %d GB de armazenamento de forma otimizada...\n", gigabytes)

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("falha ao criar arquivo de nuvem: %v", err)
	}
	defer file.Close()

	// Trunca o arquivo para o tamanho desejado (Cria um sparse file super rápido no Windows NTFS)
	sizeBytes := int64(gigabytes) * 1024 * 1024 * 1024
	if err := file.Truncate(sizeBytes); err != nil {
		return fmt.Errorf("falha ao alocar espaço: %v", err)
	}

	log.Printf("Nebula Cloud: %d GB alocados com sucesso no arquivo %s!\n", gigabytes, filename)
	return nil
}
