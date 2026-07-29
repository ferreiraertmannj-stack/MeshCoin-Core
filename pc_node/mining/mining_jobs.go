package mining

import (
	"context"
	"fmt"
	"time"
)

type MiningJob struct {
	ID        string
	Template  *BlockTemplate
	Target    string
	CreatedAt time.Time
	Ctx       context.Context
	Cancel    context.CancelFunc
}

func NewMiningJob(tmpl *BlockTemplate, target string, timeout time.Duration) *MiningJob {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	hashPrefix := tmpl.PreviousHash
	if len(hashPrefix) > 8 {
		hashPrefix = hashPrefix[:8]
	}
	id := fmt.Sprintf("job-%d-%s", tmpl.Height, hashPrefix)

	return &MiningJob{
		ID:        id,
		Template:  tmpl,
		Target:    target,
		CreatedAt: time.Time{},
		Ctx:       ctx,
		Cancel:    cancel,
	}
}

func (j *MiningJob) IsObsolete(newHeight uint64) bool {
	return j.Template.Height < newHeight
}
