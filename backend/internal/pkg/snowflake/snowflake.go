package snowflake

import (
	"sync"
	"time"
)

const (
	epoch          = int64(1704067200000) // 2024-01-01 UTC
	workerBits     = 10
	sequenceBits   = 12
	maxWorkerID    = -1 ^ (-1 << workerBits)
	maxSequence    = -1 ^ (-1 << sequenceBits)
	workerShift    = sequenceBits
	timestampShift = sequenceBits + workerBits
)

type Generator struct {
	mu        sync.Mutex
	timestamp int64
	workerID  int64
	sequence  int64
}

func New(workerID int64) *Generator {
	if workerID < 0 || workerID > maxWorkerID {
		workerID = workerID & maxWorkerID
	}
	return &Generator{workerID: workerID}
}

func (g *Generator) Next() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixMilli() - epoch
	if now == g.timestamp {
		g.sequence = (g.sequence + 1) & maxSequence
		if g.sequence == 0 {
			for now <= g.timestamp {
				now = time.Now().UnixMilli() - epoch
			}
		}
	} else {
		g.sequence = 0
	}
	g.timestamp = now
	return now<<timestampShift | g.workerID<<workerShift | g.sequence
}

func ShardKey(id int64, shardCount int) int {
	return int(id % int64(shardCount))
}
