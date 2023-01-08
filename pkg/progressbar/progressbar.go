package progressbar

import (
	"errors"
	"github.com/cheggaaa/pb"
	"github.com/mattn/go-isatty"
	"io"
	"os"
)

type Units int

const Bytes Units = 0

var AlreadyRunningErr = errors.New("progressbar has already been started")
var AlreadyStoppedErr = errors.New("progressbar has already been stopped")

type ProgressBar interface {
	Start() error
	Stop() error
	Add(total int64, units Units, prefix string) (io.Writer, error)
}

type progressBar struct {
	pool *pb.Pool
	bars []*pb.ProgressBar
}

// Start starts the progress bar pool.
func (p *progressBar) Start() error {
	var err error
	if p.pool != nil {
		return AlreadyRunningErr
	}

	p.pool, err = pb.StartPool(p.bars...)
	return err
}

// Stop stops the progress bar pool.
func (p *progressBar) Stop() error {
	if p.pool == nil {
		return AlreadyStoppedErr
	}

	if err := p.pool.Stop(); err != nil {
		return err
	}

	p.pool = nil
	return nil
}

// Add adds a progress bar to the pool before it started its execution.
func (p *progressBar) Add(total int64, units Units, prefix string) (io.Writer, error) {
	// Check if progress bar is already running.
	if p.pool != nil {
		return nil, AlreadyRunningErr
	}

	// Map units to cheggaaa's pb units.
	var pbUnit pb.Units
	if units == Bytes {
		pbUnit = pb.U_BYTES
	}

	// Create the progress bar and append it.
	bar := pb.New64(total).SetUnits(pbUnit).Prefix(prefix)
	p.bars = append(p.bars, bar)

	return bar, nil
}

type NoopProgressBar struct{}

func (n NoopProgressBar) Start() error { return nil }

func (n NoopProgressBar) Stop() error { return nil }

func (n NoopProgressBar) Add(int64, Units, string) (io.Writer, error) { return io.Discard, nil }

// NewProgressBar creates a wrapper over cheggaaa's progress bar, that simplifies the creation and execution of a
// progress bar pool. If the program is being executed outside a terminal, it returns a no-op progress bar.
func NewProgressBar() ProgressBar {
	// TODO: Check also download size > 0.
	if isatty.IsTerminal(os.Stdout.Fd()) {
		var bars []*pb.ProgressBar
		return &progressBar{bars: bars, pool: nil}
	}

	return &NoopProgressBar{}
}

var _ ProgressBar = (*progressBar)(nil)
var _ ProgressBar = (*NoopProgressBar)(nil)
