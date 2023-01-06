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
	Add(total int64, units Units, prefix string) io.Writer
}

type progressBar struct {
	pool *pb.Pool
	bars []*pb.ProgressBar
}

func (p *progressBar) Start() error {
	var err error
	if p.pool != nil {
		return AlreadyRunningErr
	}

	p.pool, err = pb.StartPool(p.bars...)
	return err
}

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

func (p *progressBar) Add(total int64, units Units, prefix string) io.Writer {
	var pbUnit pb.Units
	if units == Bytes {
		pbUnit = pb.U_BYTES
	}

	bar := pb.New64(total).SetUnits(pbUnit).Prefix(prefix)
	p.bars = append(p.bars, bar)

	return bar
}

type NoopProgressBar struct{}

func (n NoopProgressBar) Start() error { return nil }

func (n NoopProgressBar) Stop() error { return nil }

func (n NoopProgressBar) Add(int64, Units, string) io.Writer { return io.Discard }

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
