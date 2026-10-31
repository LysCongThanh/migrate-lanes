package scan

import (
	"errors"
	"fmt"

	"github.com/LysCongThanh/migrate-lanes/internal/config"
	"github.com/LysCongThanh/migrate-lanes/internal/detect"
	"github.com/LysCongThanh/migrate-lanes/internal/gitx"
	"github.com/LysCongThanh/migrate-lanes/internal/lane"
)

// Options controls one scan.
type Options struct {
	Dir     string // any directory inside the repo
	Base    string // overrides config/auto base
	Remotes bool   // also scan remote-tracking branches (CI)
}

// Result is the full scan output; it doubles as the --json schema.
type Result struct {
	Repo     string           `json:"repo"`
	Base     string           `json:"base"`
	Lanes    []*lane.Lane     `json:"lanes"`
	Targets  []detect.Target  `json:"migration_dirs"`
	Findings []detect.Finding `json:"findings"`
	baseLane *lane.Lane
}

// LaneByName returns a scanned lane (base included) or nil.
func (r *Result) LaneByName(name string) *lane.Lane {
	if r.baseLane != nil && r.baseLane.Name == name {
		return r.baseLane
	}
	for _, l := range r.Lanes {
		if l.Name == name {
			return l
		}
	}
	return nil
}

// ErrNotRepo is returned when the start directory is outside any git repo.
var ErrNotRepo = errors.New("not inside a git repository")

// Run discovers lanes and analyzes every migration directory. All errors are
// setup problems (not a repo, bad config, unknown base) -> exit code 2.
func Run(opts Options) (*Result, error) {
	root, err := gitx.TopLevel(opts.Dir)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotRepo, err)
	}
	cfg, err := config.Load(root, detect.Tools())
	if err != nil {
		return nil, err
	}
	base, err := resolveBase(root, opts.Base, cfg)
	if err != nil {
		return nil, err
	}
	baseLane, err := lane.FromRef(root, "base:"+base, lane.KindBase, base)
	if err != nil {
		return nil, err
	}
	lanes, err := discoverLanes(root, base, cfg, opts.Remotes || cfg.Remotes)
	if err != nil {
		return nil, err
	}
	targets, err := pickTargets(cfg, append([]*lane.Lane{baseLane}, lanes...))
	if err != nil {
		return nil, err
	}
	res := &Result{Repo: root, Base: base, Lanes: lanes, Targets: targets, Findings: []detect.Finding{}, baseLane: baseLane}
	for _, t := range targets {
		res.Findings = append(res.Findings, t.Detector.Analyze(t.Dir, baseLane, lanes)...)
	}
	return res, nil
}

// pickTargets uses configured paths when present, else auto-detection.
func pickTargets(cfg config.Config, all []*lane.Lane) ([]detect.Target, error) {
	if len(cfg.Paths) == 0 {
		return detect.AutoDetect(all), nil
	}
	var out []detect.Target
	for _, p := range cfg.Paths {
		d := detect.ByTool(p.Tool)
		if d == nil {
			d = detect.ClaimForDir(p.Dir, all)
		}
		if d == nil {
			return nil, fmt.Errorf("%s: cannot detect the migration tool for %q; set tool = one of %v", config.FileName, p.Dir, detect.Tools())
		}
		out = append(out, detect.Target{Dir: p.Dir, Tool: d.Tool(), Detector: d})
	}
	return out, nil
}
