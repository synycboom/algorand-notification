package fetcher

import (
	"context"
	"time"

	"github.com/algorand/go-algorand-sdk/client/v2/common"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/client/v2/indexer"
	"github.com/rs/zerolog/log"
	"go.uber.org/ratelimit"
)

// blockQueueSize is a size of buffered channel for a block queue
const blockQueueSize = 100

// ProcessorFunc represents a processor function that consume blocks
type ProcessorFunc func(*models.Block)

// Fetcher handles block fetching from algod or indexer
type Fetcher struct {
	client    *indexer.Client
	currRound uint64
	ctx       context.Context
	cancel    context.CancelFunc
	queue     chan *models.Block
	processor ProcessorFunc
	rl        ratelimit.Limiter
}

// Config represents a configuration
type Config struct {
	Host       string
	APIToken   string
	RPS        int
	StartRound *uint64
	Processor  ProcessorFunc
}

// New creates a new fetcher
func New(conf Config) (*Fetcher, error) {
	client, err := indexer.MakeClient(conf.Host, conf.APIToken)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	currRound := conf.StartRound
	if currRound == nil {
		resp, err := client.HealthCheck().Do(ctx)
		if err != nil {
			return nil, err
		}

		currRound = &resp.Round
	}

	ctx, cancel = context.WithCancel(context.Background())

	return &Fetcher{
		client:    client,
		ctx:       ctx,
		cancel:    cancel,
		currRound: *currRound,
		queue:     make(chan *models.Block, blockQueueSize),
		processor: conf.Processor,
		rl:        ratelimit.New(conf.RPS),
	}, nil
}

func (f *Fetcher) fetchLoop() {
	defer close(f.queue)

	for {
		select {
		case <-f.ctx.Done():
			return
		default:
		}

		f.rl.Take()
		nextRound := f.currRound + 1
		block, err := f.client.LookupBlock(nextRound).Do(f.ctx)
		if err != nil {
			if _, ok := err.(common.NotFound); ok {
				time.Sleep(time.Duration(1) * time.Second)
				log.Info().Msg("fetcher: no new round")

				continue
			}

			log.Error().Err(err).Msg("fetcher: got an error while looking up a block")
			continue
		}

		f.queue <- &block
		f.currRound = nextRound
		log.Info().Msgf("fetcher: current round is %v", f.currRound)
	}
}

func (f *Fetcher) processLoop() {
	for {
		select {
		case <-f.ctx.Done():
			return
		default:
		}

		block, open := <-f.queue
		if !open {
			return
		}

		if f.processor != nil {
			f.processor(block)
		}
	}
}

// Start starts fetching data from the blockchain
func (f *Fetcher) Start() {
	go f.fetchLoop()
	go f.processLoop()
}

// Stop stops the fetcher
func (f *Fetcher) Stop() {
	f.cancel()
}
